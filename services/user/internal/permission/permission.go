package permission

type Permission uint

const (
	Read   Permission = 1 << iota // 1 (001)
	Write                         // 2 (010)
	Delete                        // 4 (100)
)

func New(perm uint) Permission {
	return Permission(perm)
}

func (p Permission) String() string {
	if p == 0 {
		return "UNDEFINED"
	}
	var s string
	if p.Has(Read) {
		s += "R"
	}
	if p.Has(Write) {
		s += "W"
	}
	if p.Has(Delete) {
		s += "D"
	}
	return s
}

func (p Permission) Has(perm Permission) bool {
	return (p & perm) == perm
}
