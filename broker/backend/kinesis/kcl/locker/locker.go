package locker

type Releaser interface {
	// Release releases the underlying lock.
	Release() error
}

type Locker interface {
	// Lock creates a new distributed lock along with a Releaser that allows the
	// caler to release the lock when it's done with it. The locker must renew
	// the lock periodically if that was necessary. Lock returns an error if the
	// lock can't be created because it has already been claimed.
	Lock(string) (releaser Releaser, success bool, err error)

	// LockWait is similar to Lock but it blocks until the lock becomes
	// avaiable. The implementor may introduce a maximum wait or retries.
	LockWait(string) (releaser Releaser, err error)
}
