package main

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
)

// Stat is the avatar of file /proc/pid/stat.
// ref https://man7.org/linux/man-pages/man5/proc.5.html
type Stat struct {
	PID                 int
	Comm                string
	State               rune
	PPID                int
	PGrp                int
	Session             int
	TTYNR               int
	TPGID               int
	Flags               uint
	MinFlt              uint32
	CMinFlt             uint32
	MajFlt              uint32
	CMajFlt             uint32
	UTime               uint32
	STime               uint32
	CUTime              int32
	CSTime              int32
	Priority            int32
	Nice                int32
	NumThreads          int32
	ITRealValue         int32
	StartTime           uint64
	VSize               uint64 // Fact proved, can be maximum of uint64, as C++ use it as unsigned long (%lu) in LP64.
	RSS                 int32
	RSSLim              uint64 // unsigned long (%lu) as uint64
	StartCode           uint64 // [PT] unsigned long (%lu) as uint64
	EndCode             uint64 // [PT] unsigned long (%lu) as uint64
	StartStack          uint64 // [PT] unsigned long (%lu) as uint64
	KStkESP             uint32 // [PT]
	KStkEIP             uint32 // [PT]
	Signal              uint32
	Blocked             uint32
	SigIgnore           uint32
	SigCatch            uint32
	WChan               uint32 // [PT]
	NSwap               uint32
	CNSwap              uint32
	ExitSignal          int
	Processor           int
	RTPriority          uint
	Policy              uint
	DelayAccuBlkIOTicks uint64
	GuestTime           uint32
	CGuestTime          int32
	StartData           uint64 // [PT] unsigned long (%lu) as uint64
	EndData             uint64 // [PT] unsigned long (%lu) as uint64
	StartBrk            uint64 // [PT] unsigned long (%lu) as uint64
	ArgStart            uint64 // [PT] unsigned long (%lu) as uint64
	ArgEnd              uint64 // [PT] unsigned long (%lu) as uint64
	EnvStart            uint64 // [PT] unsigned long (%lu) as uint64
	EnvEnd              uint64 // [PT] unsigned long (%lu) as uint64
	ExitCode            int    // [PT]
}

var parseFormat string

func fieldToSpecifier(field reflect.StructField) string {
	if field.Name == "State" {
		// The ASCII rune's Type is int32, but shall use %c rather than %ld.
		return "%c"
	}
	switch field.Type.Kind() {
	case reflect.String:
		return "%s"
	case reflect.Int: // "%d"
		fallthrough
	case reflect.Uint: // "%u"
		fallthrough
	case reflect.Uint32: // "%lu"
		fallthrough
	case reflect.Int32: // "%ld"
		fallthrough
	case reflect.Uint64: // "%llu"
		// Different from that in C, golang use %d for signed and unsigned number of various width.
		return "%d"
	default:
		panic(fmt.Errorf("type unmatched field %v", field))
	}
}

func init() {
	t := reflect.TypeOf(Stat{})
	var specifiers []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		specifiers = append(specifiers, fieldToSpecifier(field))
	}
	parseFormat = strings.Join(specifiers, " ")
}

// scan reads s and output its parsed data.
// In current implementation, I avoid the usage of reflection in not-bootstrap runtime.
// There is a reflection version as scanThroughReflection in its benchmark test source file.
// On my M1 Ultra, the benchmark hints that this native version is faster as cost 10335 ns/op on sample data,
// while the reflection version is slower as 16299 ns/op. Was it worth it? Have we crossed the line as
// sacrifice the readability to chase for a little performance? BTW, the long args code are generated,
// see the initialization of specifiers as a reference of how to print every filed formatted.
func scan(s string) (*Stat, error) {
	var ret Stat
	if _, err := fmt.Sscanf(
		s,
		parseFormat,
		&ret.PID,
		&ret.Comm,
		&ret.State,
		&ret.PPID,
		&ret.PGrp,
		&ret.Session,
		&ret.TTYNR,
		&ret.TPGID,
		&ret.Flags,
		&ret.MinFlt,
		&ret.CMinFlt,
		&ret.MajFlt,
		&ret.CMajFlt,
		&ret.UTime,
		&ret.STime,
		&ret.CUTime,
		&ret.CSTime,
		&ret.Priority,
		&ret.Nice,
		&ret.NumThreads,
		&ret.ITRealValue,
		&ret.StartTime,
		&ret.VSize,
		&ret.RSS,
		&ret.RSSLim,
		&ret.StartCode,
		&ret.EndCode,
		&ret.StartStack,
		&ret.KStkESP,
		&ret.KStkEIP,
		&ret.Signal,
		&ret.Blocked,
		&ret.SigIgnore,
		&ret.SigCatch,
		&ret.WChan,
		&ret.NSwap,
		&ret.CNSwap,
		&ret.ExitSignal,
		&ret.Processor,
		&ret.RTPriority,
		&ret.Policy,
		&ret.DelayAccuBlkIOTicks,
		&ret.GuestTime,
		&ret.CGuestTime,
		&ret.StartData,
		&ret.EndData,
		&ret.StartBrk,
		&ret.ArgStart,
		&ret.ArgEnd,
		&ret.EnvStart,
		&ret.EnvEnd,
		&ret.ExitCode,
	); err != nil {
		return nil, fmt.Errorf("parse %s as format %s: %v", s, parseFormat, err)
	}
	return &ret, nil
}

// NewStat reads the stat file by pid and return the parsed data.
func NewStat(pid int) (*Stat, error) {
	bytes, err := os.ReadFile(path.Join(procPathRoot, strconv.Itoa(pid), "stat"))
	if err != nil {
		return nil, err
	}
	return scan(string(bytes))
}
