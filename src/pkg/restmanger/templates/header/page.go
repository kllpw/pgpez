package header

type PageData struct {
	PageTitle string
	KeysActive bool
	MessagesActive bool
	ContactsActive bool
	DarkMode bool
}

func (pd *PageData) GetData() interface{} {
	return pd
}


