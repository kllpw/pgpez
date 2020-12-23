package restmanger

import (
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/kllpw/pgpez/src/pkg/contacts"
	"github.com/kllpw/pgpez/src/pkg/processor"
	"github.com/kllpw/pgpez/src/pkg/restmanger/templates"
	templateContacts "github.com/kllpw/pgpez/src/pkg/restmanger/templates/contacts"
	templateErr "github.com/kllpw/pgpez/src/pkg/restmanger/templates/error"
	templateIndex "github.com/kllpw/pgpez/src/pkg/restmanger/templates/index"
	templateKeys "github.com/kllpw/pgpez/src/pkg/restmanger/templates/keys"
	templateMsg "github.com/kllpw/pgpez/src/pkg/restmanger/templates/messages"
	"net/http"
	"os"
)

var (
	ErrValidationFailed = errors.New("validation failed")
)

type Manager struct {
	processor processor.ProcReqs
	renderer  templates.Renderer
}

func NewManager() *Manager {
	dir, _ := os.Getwd()
	pro, _ := processor.NewProcessor(dir + "/pgpez.db")
	mngr := &Manager{
		processor: pro,
		renderer:  templates.DefaultRenderer,
	}
	r := mux.NewRouter()

	r.NewRoute().Name("index").Path("/").HandlerFunc(mngr.indexHandler)
	dm := r.NewRoute().Name("dm").PathPrefix("/dm").Subrouter()
	dm.HandleFunc("", mngr.darkModeHandler)

	keyR := r.NewRoute().Name("keys").PathPrefix("/keys").Subrouter()
	keyR.HandleFunc("", mngr.keysHandler)
	keyR.HandleFunc("/{name}", mngr.keysHandler)

	contactR := r.NewRoute().Name("contacts").PathPrefix("/contacts").Subrouter()
	contactR.HandleFunc("", mngr.contactsAllHandler)

	msgR := r.NewRoute().Name("messages").PathPrefix("/messages").Subrouter()
	msgR.HandleFunc("", mngr.messagesAllHandler)
	msgR.HandleFunc("/decrypt", mngr.decryptMessageHandler)
	msgR.HandleFunc("/encrypt", mngr.encryptMessageHandler)

	http.Handle("/", r)
	return mngr
}
func (m *Manager) Start() error {
	return http.ListenAndServe(":80", nil)
}

func (m *Manager) messagesAllHandler(w http.ResponseWriter, r *http.Request) {
	kys, err := m.processor.GetAllKeys()
	c, err := m.processor.GetAllContacts()
	m.CheckError(w, err)
	m.renderer.RenderTemplate(w, templates.Messages, &templateMsg.PageData{
		PageTitle: "",
		Contacts:  c,
		Keys:      kys,
		Message:   "",
	})
}

func (m *Manager) encryptMessageHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("encryptMessageHandler")
	id := r.FormValue("id")
	message := r.FormValue("emessage")
	res, err := m.processor.EncryptMessage(id, message, true)
	m.CheckError(w, err)
	kys, err := m.processor.GetAllKeys()
	m.CheckError(w, err)
	c, err := m.processor.GetAllContacts()
	m.CheckError(w, err)
	m.renderer.RenderTemplate(w, templates.Messages, &templateMsg.PageData{
		PageTitle: "",
		Contacts:  c,
		Keys:      kys,
		Message:   res,
	})
}

func (m *Manager) keyDeleteHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("keyDeleteHandler")
	id := r.FormValue("id")
	passphrase := r.FormValue("passphrase")
	err := m.processor.DeleteKey(id, passphrase)
	m.CheckError(w, err)
	http.Redirect(w, r, "/keys", http.StatusSeeOther)
	return
}

func (m *Manager) decryptMessageHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("decryptMessageHandler")
	id := r.FormValue("id")
	passphrase := r.FormValue("passphrase")
	message := r.FormValue("message")
	res, err := m.processor.DecryptMessage(id, passphrase, message)
	ok := m.CheckError(w, err)
	kys, err := m.processor.GetAllKeys()
	ok = ok && m.CheckError(w, err)
	c, err := m.processor.GetAllContacts()
	ok = ok && m.CheckError(w, err)
	if ok {
		m.renderer.RenderTemplate(w, templates.Messages, &templateMsg.PageData{
			PageTitle: "decrypted message",
			Message:   res,
			Keys:      kys,
			Contacts:  c,
		})
	}
	return
}

func (m *Manager) keysNameHandler(w http.ResponseWriter, r *http.Request, name string, locked bool) {
	fmt.Println("KeysNameHandler")
	k, err := m.processor.GetKeyByName(name)
	ok := m.CheckError(w, err)
	if ok {
		m.renderer.RenderTemplate(w, templates.Key, &templateKeys.PageData{
			PageTitle: "Key",
			KeyCount:  1,
			Key:       k,
			Locked:    locked,
		})
	}
	return
}

func (m *Manager) keyAddHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("KeysAddHandler")
	name := r.FormValue("name")
	passphrase := r.FormValue("passphrase")
	ok := name != "" && passphrase != ""
	if !ok {
		m.CheckError(w, ErrValidationFailed)
	}
	_, err := m.processor.GenerateAndStoreNewKeyPair(name, passphrase)
	ok = ok && m.CheckError(w, err)
	if ok {
		http.Redirect(w, r, "/keys", http.StatusSeeOther)
	}
}

func (m *Manager) keysHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("KeysAllHandler")
	if r.Method == http.MethodPost {
		action := r.FormValue("action")
		switch action {
		case http.MethodPost:
			fmt.Println("KeysAddHandler")
			m.keyAddHandler(w, r)
			return
		case http.MethodDelete:
			fmt.Println("KeysDeleteHandler")
			m.keyDeleteHandler(w, r)
			return
		case http.MethodPatch:
			fmt.Println("KeysDeleteHandler")
			m.keyAuthHandler(w, r)
			return
		}
	} else if r.Method == http.MethodGet {
		vars := mux.Vars(r)
		name := vars["name"]
		if name != "" {
			m.keysNameHandler(w, r, name, true)
			return
		}
		ks, err := m.processor.GetAllKeys()
		ok := m.CheckError(w, err)
		if ok {
			m.renderer.RenderTemplate(w, templates.Keys, &templateKeys.PageData{
				PageTitle: "Keys",
				KeyCount:  len(ks),
				Keys:      ks,
			})
		}
	}
}

func (m *Manager) indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("IndexHandler")
	m.renderer.RenderTemplate(w, templates.Index, &templateIndex.PageData{
		PageTitle:      "keystore",
		WelcomeMessage: "Welcome to the keystore",
	})
}

func (m *Manager) contactsAllHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("ContactsAllHandler")
	if r.Method == http.MethodPost {
		action := r.FormValue("action")
		switch action {
		case http.MethodPost:
			fmt.Println("ContactsAddHandler")
			m.contactAddHandler(w, r)
			return
		case http.MethodDelete:
			fmt.Println("ContactsDeleteHandler")
			m.contactDeleteHandler(w, r)
			return
		}
	} else if r.Method == http.MethodGet {
		vars := mux.Vars(r)
		name := vars["name"]
		if name != "" {
			m.contactNameHandler(w, r, name)
		}
		contacts, err := m.processor.GetAllContacts()
		m.CheckError(w, err)
		m.renderer.RenderTemplate(w, templates.Contacts,
			&templateContacts.PageData{
				PageTitle:    "Contacts",
				ContactCount: len(contacts),
				Contacts:     contacts,
			})
	}
}

func (m *Manager) CheckError(w http.ResponseWriter, err error) bool {
	if err != nil {
		m.renderer.RenderTemplate(w, templates.Error,
			&templateErr.PageData{
				PageTitle: "Error",
				Error:     err.Error(),
			})
		return false
	}
	return true
}

func (m *Manager) contactAddHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("ContactAddHandler")
	name := r.FormValue("name")
	pubkey := r.FormValue("pubkey")
	ok := name != "" && pubkey != ""
	if !ok {
		m.CheckError(w, ErrValidationFailed)
		return
	}
	_, err := m.processor.AddContact(name, pubkey)
	if err != nil {
		m.CheckError(w, err)
		return
	}
	http.Redirect(w, r, "/contacts", http.StatusSeeOther)

}

func (m *Manager) contactDeleteHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("ContactDeleteHandler")
	id := r.FormValue("id")
	err := m.processor.DeleteContact(id)
	if err != nil {
		m.CheckError(w, err)
		return
	}
	http.Redirect(w, r, "/contacts", http.StatusSeeOther)
	return
}

func (m *Manager) contactNameHandler(w http.ResponseWriter, r *http.Request, name string) {
	c, err := m.processor.GetContactByName(name)
	m.CheckError(w, err)
	m.renderer.RenderTemplate(w, templates.Contacts,
		&templateContacts.PageData{
			PageTitle:    "Contacts",
			ContactCount: 1,
			Contacts:     []*contacts.Contact{c},
		})
}

func (m *Manager) keyAuthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("keyAuthHandler")
	id := r.FormValue("id")
	passphrase := r.FormValue("passphrase")
	locked := true
	err := m.processor.AuthKey(id, passphrase)
	if err == nil {
		locked = false
	}
	k, err := m.processor.GetKeyById(id)
	r, err = http.NewRequest("GET", "/keys/"+k.Name, http.NoBody)
	m.keysNameHandler(w, r, k.Name, locked)
}

func (m *Manager) darkModeHandler(w http.ResponseWriter, r *http.Request) {
	m.renderer.ToggleDarkMode()
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
