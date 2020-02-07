package policystorage

import "time"

type PolicyStorage interface {
	List() ([]*PolicyListStub, error)
	Notify() (<-chan []*PolicyListStub, <-chan error)
	Get(string) (*Policy, error)
}

type Policy struct {
	ID       string
	Source   string
	Query    string
	Interval time.Duration
	Target   *Target
	Strategy *Strategy
}

type PolicyListStub struct {
	ID     string
	Source string
	Query  string
	Target
	Strategy
}

type Strategy struct {
	Name   string
	Min    int
	Max    int
	Config map[string]string
}

type Target struct {
	Name   string
	Config map[string]string
}
