package main

import (
	"flag"
	"fmt"
	"github.com/libgit2/git2go"
	"os"
	"time"
)

const (
	ModeNone = iota
	ModePrefix
	ModeRepeat
)

var commitFormat string = "tree %s\nparent %s\nauthor %s <%s> %d %+05d\ncommitter %[3]s <%[4]s> %[5]d %+05[6]d\n\n%s\n\nNovelty salt: %s\n"

func hexStringToByteSlices(s string) (match []byte, mask []byte, err error) {
	match = make([]byte, 20)
	mask = make([]byte, 20)

	var x byte
	for i, c := range s {
		switch c {
		case '0':
			x = 0
		case '1':
			x = 1
		case '2':
			x = 2
		case '3':
			x = 3
		case '4':
			x = 4
		case '5':
			x = 5
		case '6':
			x = 6
		case '7':
			x = 7
		case '8':
			x = 8
		case '9':
			x = 9
		case 'a':
			x = 10
		case 'b':
			x = 11
		case 'c':
			x = 12
		case 'd':
			x = 13
		case 'e':
			x = 14
		case 'f':
			x = 15
		default:
			err = fmt.Errorf("illegal character: %v", c)
			break
		}

		if i%2 == 0 {
			match[i/2] = x << 4
			mask[i/2] = 0xf0
		} else {
			match[i/2] |= x
			mask[i/2] |= 0x0f
		}
	}

	return
}

var digits94 string = "!\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~"

func base94(i int) string {
	var b []byte
	for j := i; j > 0; j /= 94 {
		b = append([]byte{digits94[j%94]}, b...)
	}
	return string(b)
}

func main() {
	var (
		msg         string
		strPrefix   string
		strRepeat   string
		cycle       uint
		err         error
		match, mask []byte
		mode        int
	)

	flag.StringVar(&msg, "m", "", "commit message")
	flag.StringVar(&strPrefix, "p", "", "desired prefix")
	flag.StringVar(&strRepeat, "r", "", "desired repeat")
	flag.UintVar(&cycle, "c", 5, "cycle for repeating")
	flag.Parse()

	if msg == "" {
		err = fmt.Errorf("message is required")
	} else {
		if strPrefix != "" {
			if len(strPrefix) >= 40 {
				err = fmt.Errorf("prefix is too long")
			} else {
				mode = ModePrefix
			}
		}
		if err == nil && strRepeat != "" {
			if mode != ModeNone {
				err = fmt.Errorf("prefix and repeat settings are mutually exclusive")
				mode = ModeNone
			} else if cycle >= 40 || cycle == 0 {
				err = fmt.Errorf("cycle must be between 1 and 40 (inclusive)")
			} else {
				mode = ModeRepeat
			}
		}
		if err == nil && mode == ModeNone {
			err = fmt.Errorf("prefix or repeat is required")
		}
	}

	switch mode {
	case ModePrefix:
		match, mask, err = hexStringToByteSlices(strPrefix)
	case ModeRepeat:
		match, mask, err = hexStringToByteSlices(strRepeat)

		if err == nil {
			for i, j := cycle, 0; i < 40; i++ {
				if i%cycle == 0 {
					j = 0
				}
				if j < len(strRepeat) {
					var c byte
					if j%2 == 0 {
						c = match[j/2] >> 4
					} else {
						c = match[j/2] & 0x0f
					}

					if i%2 == 0 {
						match[i/2] = c<<4 + match[i/2]&0x0f
						mask[i/2] |= 0xf0
					} else {
						match[i/2] = c + match[i/2]&0xf0
						mask[i/2] |= 0x0f
					}
					j++
				}
			}
		}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\nUsage: %s -m <message> (-p <prefix>|-r <repeat> -c <cycle>)\n", err, os.Args[0])
		return
	}

	configfn, err := git.ConfigFindGlobal()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error finding global config file:", err)
		return
	}

	config, err := git.NewConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating config:", err)
		return
	}

	config, err = git.OpenOndisk(config, configfn)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error loading config:", err)
		return
	}

	user, err := config.LookupString("user.name")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error looking up user.name:", err)
		return
	}

	email, err := config.LookupString("user.email")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error looking up user.email:", err)
		return
	}

	repo, err := git.OpenRepository(".")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error opening repository:", err)
		return
	}

	head, err := repo.Head()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error fetching HEAD:", err)
		return
	}

	odb, err := repo.Odb()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error fetching object database:", err)
		return
	}

	index, err := repo.Index()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error fetching index:", err)
		return
	}

	tree, err := index.WriteTree()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error writing tree:", err)
		return
	}

	now := time.Now()
	_, offset := now.Zone()

	var (
		content string
		oid     *git.Oid
		i       int
		ok      bool
	)
	for {
		if i%100000 == 99999 {
			fmt.Fprintln(os.Stderr, "Attempt:", i+1)
		}

		content = fmt.Sprintf(commitFormat, tree, head.Target(), user, email, now.Unix(), offset/36, msg, base94(i))
		oid, err = odb.Hash([]byte(content), git.ObjectCommit)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error hashing object:", err)
			return
		}

		ok = true
		for i, c := range oid {
			if c&mask[i] != match[i] {
				ok = false
				break
			}
		}
		if ok {
			break
		}

		i++
	}

	oid, err = odb.Write([]byte(content), git.ObjectCommit)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error writing object:", err)
		return
	}

	commit, err := repo.LookupCommit(oid)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error looking up commit:", err)
		return
	}

	opts := new(git.CheckoutOpts)
	err = repo.ResetToCommit(commit, git.ResetMixed, opts)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error resetting to commit:", err)
		return
	}

	fmt.Fprintln(os.Stderr, "Result:", oid, "Attempts:", i+1)
}
