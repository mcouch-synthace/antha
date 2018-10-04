package sampletracker

import (
	"sync"

	"github.com/antha-lang/antha/antha/anthalib/wtype"
)

var stLock sync.Mutex
var st *SampleTracker

// SampleTracker record the location of components generated during element execution
// as well as any explicitly set input plates
type SampleTracker struct {
	lock     sync.Mutex
	records  map[string]string
	forwards map[string]string
	plates   map[string]*wtype.Plate
}

func newSampleTracker() *SampleTracker {
	return &SampleTracker{
		records:  make(map[string]string),
		forwards: make(map[string]string),
		plates:   make(map[string]*wtype.Plate),
	}
}

func GetSampleTracker() *SampleTracker {
	stLock.Lock()
	defer stLock.Unlock()

	if st == nil {
		st = newSampleTracker()
	}

	return st
}

// SetInputPlate declare the given plate as an input to the experiment
// recording the id and location of every sample in it
func (st *SampleTracker) SetInputPlate(p *wtype.Plate) {
	st.lock.Lock()
	defer st.lock.Unlock()

	st.plates[p.ID] = p

	for _, w := range p.HWells {
		if !w.IsEmpty() {
			st.setLocationOf(w.WContents.ID, w.WContents.Loc)
			w.SetUserAllocated()
		}
	}
}

// GetInputPlates return a list of all input plates explicitly set during element
// execution
func (st *SampleTracker) GetInputPlates() []*wtype.Plate {
	st.lock.Lock()
	defer st.lock.Unlock()

	var ret []*wtype.Plate
	if len(st.plates) == 0 {
		return ret
	}
	ret = make([]*wtype.Plate, 0, len(st.plates))

	for _, p := range st.plates {
		ret = append(ret, p)
	}

	st.plates = make(map[string]*wtype.Plate)

	return ret
}

func (st *SampleTracker) setLocationOf(ID string, loc string) {
	st.records[ID] = loc
}

// SetLocationOf set the string encoded location of the component with the given ID
func (st *SampleTracker) SetLocationOf(ID string, loc string) {
	st.lock.Lock()
	defer st.lock.Unlock()
	st.setLocationOf(ID, loc)
}

func (st *SampleTracker) getLocationOf(ID string) (string, bool) {
	if ID == "" {
		return "", false
	}

	s, ok := st.records[ID]

	// look to see if there's a forwarding address
	// can this lead to an out of date location???

	if !ok {
		return st.getLocationOf(st.forwards[ID])
	}

	return s, ok
}

// GetLocationOf return the string location of the component with the given ID.
// If no such component is known, the returned location will be the empty string
func (st *SampleTracker) GetLocationOf(ID string) (string, bool) {
	st.lock.Lock()
	defer st.lock.Unlock()

	return st.getLocationOf(ID)
}

// UpdateIDOf add newID as an alias for ID, such that both refer to the same location
func (st *SampleTracker) UpdateIDOf(ID string, newID string) {
	st.lock.Lock()
	defer st.lock.Unlock()
	_, ok := st.records[ID]
	if ok {
		st.records[newID] = st.records[ID]
	} else {
		// set up a forward
		// actually a backward...
		st.forwards[newID] = ID
	}
}
