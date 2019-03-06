//+build manual linux

package metric

import (
	"context"
	"github.com/orbs-network/orbs-network-go/instrumentation/log"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestSystemMetrics(t *testing.T) {
	//stat, _ := linux.ReadProcessStat("../../vendor/github.com/c9s/goprocinfo/linux/proc/3323/stat")
	//fmt.Println(stat.Rss)
	//
	//

	go func() {
		ioutil.ReadFile("/dev/random")
	}()

	m := NewRegistry()
	l := log.GetLogger().WithOutput(log.NewFormattingOutput(os.Stderr, log.NewHumanReadableFormatter()))
	NewSystemReporter(context.Background(), m, l)
	m.PeriodicallyReport(context.Background(), l)

	<-time.After(1 * time.Minute)
}