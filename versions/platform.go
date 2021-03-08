package versions

type Platform struct {
	Os   string `json:"os"`
	Arch string `json:"arch"`
}

func (p *Platform) Equals(o *Platform) bool {
	return p.Os == o.Os && p.Arch == o.Arch
}

type PlatformList []Platform

func (a PlatformList) Len() int      { return len(a) }
func (a PlatformList) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a PlatformList) Less(i, j int) bool {
	return a[i].Os < a[j].Os || a[i].Os == a[j].Os && a[i].Arch < a[j].Arch
}
