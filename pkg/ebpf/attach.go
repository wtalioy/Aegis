package ebpf

import (
	"fmt"
	"log"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
)

type lsmHook struct {
	name    string
	program **ebpf.Program
}

func AttachLSMHooks(objs *LSMObjects) ([]link.Link, error) {
	hooks := []lsmHook{
		{"bprm_check_security", &objs.LsmBprmCheck},
		{"file_open", &objs.LsmFileOpen},
		{"socket_connect", &objs.LsmSocketConnect},
	}

	var links []link.Link
	for _, h := range hooks {
		if *h.program == nil {
			continue
		}
		l, err := link.AttachLSM(link.LSMOptions{
			Program: *h.program,
		})
		if err != nil {
			CloseLinks(links)
			return nil, fmt.Errorf("attach %s LSM: %w", h.name, err)
		}
		links = append(links, l)
	}

	log.Printf("Attached %d BPF LSM hooks for active defense", len(links))
	return links, nil
}

func CloseLinks(links []link.Link) {
	for _, l := range links {
		_ = l.Close()
	}
}
