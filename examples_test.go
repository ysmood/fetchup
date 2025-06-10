package fetchup_test

import (
	"fmt"

	"github.com/ysmood/fetchup/pkg/golang_migrate"
	"github.com/ysmood/fetchup/pkg/golangci_lint"
)

func Example_install_golangci_lint() {
	err := golangci_lint.Install()
	if err != nil {
		panic(err)
	}

	fmt.Print(shellExec("./bin/golangci-lint version"))

	// Output:
	// golangci-lint has version 2.1.6 built with go1.24.2 from eabc2638 on 2025-05-04T15:41:19Z
}

func Example_install_golang_migrate() {
	err := golang_migrate.Install()
	if err != nil {
		panic(err)
	}

	fmt.Print(shellExec("./bin/migrate -version"))

	// Output:
	// 4.18.3
}
