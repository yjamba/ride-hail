package ports

type Notifier interface {
	Notify(event interface{}) error
}
