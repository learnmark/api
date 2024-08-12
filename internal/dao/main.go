package dao

// Interface is the interface for learnmark.
type Interface interface {
	GeneralDao() GeneralDao
	UserDao() UserDao
}
