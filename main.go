package main

import (
	"flag"
	"fmt"
	"github.com/libgit2/git2go"
	"os"
	"time"
)

var commitFormat string = "tree %s\nparent %s\nauthor %s <%s> %d %+05d\ncommitter %[3]s <%[4]s> %[5]d %+05[6]d\n\n%s\n\nNovelty salt: %x\n"

func main() {
	var (
		msg, strPrefix string
		err            error
		prefix, mask   []byte
	)

	flag.StringVar(&msg, "m", "", "commit message")
	flag.StringVar(&strPrefix, "p", "beef", "desired hash prefix")
	flag.Parse()

	if msg == "" || strPrefix == "" {
		err = fmt.Errorf("message and prefix are required")
	} else if len(strPrefix) >= 40 {
		err = fmt.Errorf("prefix is too long")
	} else {
		var x, y byte
		for i, c := range strPrefix {
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
				y = x << 4
			} else {
				y |= x
				prefix = append(prefix, y)
				mask = append(mask, 0xff)
			}
		}
		if len(strPrefix)%2 == 1 {
			prefix = append(prefix, y)
			mask = append(mask, 0xf0)
		}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\nUsage: %s -m <message> [-p <prefix>]\n", err, os.Args[0])
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

		content = fmt.Sprintf(commitFormat, tree, head.Target(), user, email, now.Unix(), offset/36, msg, i)
		oid, err = odb.Hash([]byte(content), git.ObjectCommit)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error hashing object:", err)
			return
		}

		ok = true
		for i, p := range prefix {
			if oid[i]&mask[i] != p {
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
