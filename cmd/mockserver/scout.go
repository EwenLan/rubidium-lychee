package main

// hasOwnScoutMarker reports whether the node has any active own scout marker.
func (e *engine) hasOwnScoutMarker(nodeID string) bool {
	markers, ok := e.scoutMarkers[nodeID]
	return ok && len(markers) > 0
}

// consumeScoutMarker consumes the earliest own scout marker at nodeID (FIFO).
// Returns true if a marker was consumed.
func (e *engine) consumeScoutMarker(nodeID string) bool {
	markers, ok := e.scoutMarkers[nodeID]
	if !ok || len(markers) == 0 {
		return false
	}
	e.scoutMarkers[nodeID] = markers[1:]
	if len(e.scoutMarkers[nodeID]) == 0 {
		delete(e.scoutMarkers, nodeID)
	}
	return true
}

// expireScoutMarkers removes markers past their expiry frame.
func (e *engine) expireScoutMarkers() {
	for nodeID, markers := range e.scoutMarkers {
		var kept []scoutMarker
		for _, m := range markers {
			if e.round <= m.ExpiryFrame {
				kept = append(kept, m)
			}
		}
		if len(kept) == 0 {
			delete(e.scoutMarkers, nodeID)
		} else {
			e.scoutMarkers[nodeID] = kept
		}
	}
}
