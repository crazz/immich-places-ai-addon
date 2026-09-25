package main

import "immich-places-backend/internal/ai/writeback"

func (s *aiWriteStore) enter(op writeback.Operation) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.active) >= 2 {
		return false
	}
	for target, owner := range s.active {
		if target == op.Plan.TargetID || owner == op.Plan.Owner {
			return false
		}
	}
	if s.active == nil {
		s.active = make(map[string]string)
	}
	s.active[op.Plan.TargetID] = op.Plan.Owner
	return true
}
func (s *aiWriteStore) leave(op writeback.Operation) {
	s.mu.Lock()
	delete(s.active, op.Plan.TargetID)
	s.mu.Unlock()
}
