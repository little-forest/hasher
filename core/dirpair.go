package core

type DirPairStatus uint8

const (
	BASE_ONLY = iota + 1
	PAIR
	TARGET_ONLY
)

type DirPair struct {
	Base   *DirDiff
	Target *DirDiff
	Status DirPairStatus
}

func NewDirPair(base *DirDiff, target *DirDiff) *DirPair {
	return &DirPair{
		Base:   base,
		Target: target,
		Status: PAIR,
	}
}

func NewBaseOnlyDirPair(base *DirDiff) *DirPair {
	return &DirPair{
		Base:   base,
		Status: BASE_ONLY,
	}
}

func NewTargetOnlyDirPair(target *DirDiff) *DirPair {
	return &DirPair{
		Target: target,
		Status: TARGET_ONLY,
	}
}

func (d DirPair) Path() string {
	if d.Base != nil {
		return d.Base.Path
	} else if d.Target != nil {
		return d.Target.Path
	} else {
		return ""
	}
}
