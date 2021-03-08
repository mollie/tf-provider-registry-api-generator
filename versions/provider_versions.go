package versions

import (
	"log"
	"sort"
	"strconv"
	"strings"
)

type ProviderVersion struct {
	Version   string     `json:"version"`
	Protocols []string   `json:"protocols"`
	Platforms []Platform `json:"platforms"`
}

type ProviderVersionList []ProviderVersion

type SemVer []int

func MakeSemVerFromString(semver string) SemVer {
	var result SemVer
	parts := strings.Split(semver, ".")

	for _, v := range parts {
		value, err := strconv.Atoi(v)
		if err != nil {
			log.Fatalf("invalid semver '%s'", semver)
		}
		result = append(result, value)
	}

	return result
}

func (v ProviderVersion) GetSemVer() SemVer {
	result := MakeSemVerFromString(v.Version)
	if len(result) != 3 {
		log.Fatalf("invalid semantic version '%s'", v.Version)
	}
	return result
}

func (v SemVer) Less(o SemVer) bool {
	var i int
	for i, _ = range v {
		if i < len(o) && v[i] != o[i] {
			return v[i] < o[i]
		}
	}

	return len(v) < len(o)
}

func (a ProviderVersionList) Len() int           { return len(a) }
func (a ProviderVersionList) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ProviderVersionList) Less(i, j int) bool { return a[i].GetSemVer().Less(a[j].GetSemVer()) }

type ProtocolList []string
func (a ProtocolList) Len() int      { return len(a) }
func (a ProtocolList) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ProtocolList) Less(i, j int) bool {
	return MakeSemVerFromString(a[i]).Less(MakeSemVerFromString(a[j]))
}

func (l *ProviderVersions) Add(meta *BinaryMetaData) {

	if l.Versions == nil {
		l.Versions = make([]ProviderVersion, 0)
	}
	for i, v := range l.Versions {
		if v.Version == meta.Version {
			l.Versions[i].AddProtocols(meta.Protocols)
			l.Versions[i].AddPlatform(Platform{Os: meta.Os, Arch: meta.Arch})
			return
		}
	}

	version := ProviderVersion{Version: meta.Version}
	version.AddProtocols(meta.Protocols)
	version.AddPlatform(Platform{Os: meta.Os, Arch: meta.Arch})
	l.Versions = append(l.Versions, version)
	sort.Sort(ProviderVersionList(l.Versions))
}

func (v *ProviderVersion) AddProtocol(protocol string) {
	if v.Protocols == nil {
		v.Protocols = make([]string, 0)
	}
	for _, p := range v.Protocols {
		if p == protocol {
			return // protocol already in list
		}
	}
	v.Protocols = append(v.Protocols, protocol)
	sort.Sort(ProtocolList(v.Protocols))
}

func (v *ProviderVersion) AddProtocols(protocols []string) {
	for _, p := range protocols {
		v.AddProtocol(p)
	}
}

func (v *ProviderVersion) AddPlatform(platform Platform) {
	if v.Platforms == nil {
		v.Platforms = make([]Platform, 0)
	}
	for _, p := range v.Platforms {
		if p.Equals(&platform) {
			return // platform already in list
		}
	}
	v.Platforms = append(v.Platforms, platform)
	sort.Sort(PlatformList(v.Platforms))
}

func (v *ProviderVersion) AddPlatforms(platforms []Platform) {
	for _, platform := range platforms {
		v.AddPlatform(platform)
	}
}

type ProviderVersions struct {
	Versions []ProviderVersion `json:"versions"`
}

func (p *ProviderVersions) FindVersion(version string) *ProviderVersion {
	if p.Versions == nil {
		return nil
	}
	for i, v := range p.Versions {
		if v.Version == version {
			return &p.Versions[i]
		}
	}
	return nil
}

func (p *ProviderVersions) AddProviderVersion(v ProviderVersion) {
	if p.Versions == nil {
		p.Versions = make([]ProviderVersion, 0)
	}

	p.Versions = append(p.Versions, v)
	sort.Sort(ProviderVersionList(p.Versions))
}

func (p *ProviderVersions) AddOrUpdateProviderVersion(v ProviderVersion) {
	if p.Versions == nil {
		p.AddProviderVersion(v)
		return
	}
	existingVersion := p.FindVersion(v.Version)
	if existingVersion == nil {
		p.AddProviderVersion(v)
		return
	}

	existingVersion.AddProtocols(v.Protocols)
	existingVersion.AddPlatforms(v.Platforms)
}

func (p *ProviderVersions) Merge(o ProviderVersions) {
	if o.Versions == nil {
		return
	}
	for _, version := range o.Versions {
		p.AddOrUpdateProviderVersion(version)
	}
}
