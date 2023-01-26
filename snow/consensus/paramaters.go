package consensus

import "fmt"

type Parameters struct {
	K     int
	Alpha int
	Beta  int
}

// Verify returns nil if the parameters describe a valid initialization.
func (p *Parameters) Verify() error {
	switch {
	case p.Alpha <= p.K/2:
		return fmt.Errorf("k = %d, alpha = %d: fails the condition that: k/2 < alpha", p.K, p.Alpha)
	case p.K < p.Alpha:
		return fmt.Errorf("k = %d, alpha = %d: fails the condition that: alpha <= k", p.K, p.Alpha)
	default:
		return nil
	}
}
