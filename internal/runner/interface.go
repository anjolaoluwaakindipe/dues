package runner

import (
	"os"
	"sync"
)

type Runner interface {
  CommandLoop(wg *sync.WaitGroup, sigs chan os.Signal) 
}
