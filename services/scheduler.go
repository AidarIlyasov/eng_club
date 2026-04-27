package services

import (
	sh "github.com/carlescere/scheduler"
)

func Schedule(seconds int, f func(), notImmediately bool) {
	if notImmediately {
		_, _ = sh.Every(seconds).Seconds().NotImmediately().Run(f)
	} else {
		_, _ = sh.Every(seconds).Seconds().Run(f)
	}
}
