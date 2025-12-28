package ebpf

import (
	"fmt"
	"log"

	"github.com/cilium/ebpf/link"
)

func AttachLSMHooks(objs *LSMObjects) ([]link.Link, error) {
	var links []link.Link

	lsmBprm, err := link.AttachLSM(link.LSMOptions{
		Program: objs.LsmBprmCheck,
	})
	if err != nil {
		return nil, fmt.Errorf("attach bprm_check_security LSM: %w", err)
	}
	links = append(links, lsmBprm)

	lsmFileOpen, err := link.AttachLSM(link.LSMOptions{
		Program: objs.LsmFileOpen,
	})
	if err != nil {
		CloseLinks(links)
		return nil, fmt.Errorf("attach file_open LSM: %w", err)
	}
	links = append(links, lsmFileOpen)

	lsmSocketConnect, err := link.AttachLSM(link.LSMOptions{
		Program: objs.LsmSocketConnect,
	})
	if err != nil {
		CloseLinks(links)
		return nil, fmt.Errorf("attach socket_connect LSM: %w", err)
	}
	links = append(links, lsmSocketConnect)

	log.Printf("Attached 3 BPF LSM hooks for active defense")
	return links, nil
}

func CloseLinks(links []link.Link) {
	for _, l := range links {
		_ = l.Close()
	}
}
