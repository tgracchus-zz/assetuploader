package scheduler


type Store interface {
	Add(job *Job) error
}