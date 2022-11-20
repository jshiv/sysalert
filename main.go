package main

import (
	"fmt"
	"log"
	"net/smtp"
	"strconv"
	"strings"

	human "github.com/dustin/go-humanize"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/spf13/viper"
)

type Mail struct {
	Sender  string
	To      []string
	Subject string
	Body    string
}

func BuildMessage(mail Mail) string {
	msg := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\r\n"
	msg += fmt.Sprintf("From: %s\r\n", mail.Sender)
	msg += fmt.Sprintf("To: %s\r\n", strings.Join(mail.To, ";"))
	msg += fmt.Sprintf("Subject: %s\r\n", mail.Subject)
	msg += fmt.Sprintf("\r\n%s\r\n", mail.Body)

	return msg
}

type SysStats struct {
	DiskStat *disk.UsageStat
	MemStat  *mem.VirtualMemoryStat
}

func (sys *SysStats) GetDiskStat(path string) (string, error) {
	usage, err := disk.Usage(path)
	if err != nil {
		return "", err
	}
	sys.DiskStat = usage
	percent := fmt.Sprintf("%2.f%%", usage.UsedPercent)

	formatter := "Path: %s\n Total: %s\n Used: %s\n, Free: %s\n Pct: %s\n"
	str := fmt.Sprintf(formatter,
		path,
		human.Bytes(usage.Total),
		human.Bytes(usage.Used),
		human.Bytes(usage.Free),
		percent,
	)
	return str, nil
}

func printUsage() string {
	formatter := "%-14s %7s %7s %7s %4s %s\n"
	fmt.Printf(formatter, "Filesystem", "Size", "Used", "Avail", "Use%", "Mounted on")

	parts, _ := disk.Partitions(true)
	str := ""
	for _, p := range parts {
		device := p.Mountpoint
		s, _ := disk.Usage(device)

		if s.Total == 0 {
			continue
		}

		percent := fmt.Sprintf("%2.f%%", s.UsedPercent)

		str += fmt.Sprintf(formatter,
			s.Fstype,
			human.Bytes(s.Total),
			human.Bytes(s.Used),
			human.Bytes(s.Free),
			percent,
			p.Mountpoint,
		)
	}
	return str
}

func main() {

	// any approach to require this configuration into your program.

	viper.SetConfigName("sysalert")        // name of config file (without extension)
	viper.SetConfigType("yaml")            // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(".")               // optionally look for config in the working directory
	viper.AddConfigPath("/etc/sysalert/")  // path to look for the config file in
	viper.AddConfigPath("$HOME/.sysalert") // call multiple times to add many search paths
	err := viper.ReadInConfig()            // Find and read the config file
	if err != nil {                        // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	// Configuration
	smtpConfig := viper.GetStringMapString("smtp")
	from := smtpConfig["from"]
	password := smtpConfig["password"]
	to := []string{smtpConfig["to"]}
	smtpHost := smtpConfig["host"]
	smtpPort := smtpConfig["port"]

	// usage := printUsage()
	// fmt.Println(usage)

	diskConfig := viper.GetStringMapString("disk")
	sysStat := SysStats{}
	diskStatStr, err := sysStat.GetDiskStat(diskConfig["path"])
	fmt.Println(diskStatStr)
	if err != nil {
		log.Fatal(err)
	}
	diskAlertThreshold, err := strconv.ParseFloat(diskConfig["pct"], 64)
	if err != nil {
		log.Fatal(err)
	}
	if sysStat.DiskStat.UsedPercent > diskAlertThreshold {
		subject := "Disk Threshold Alert"
		body := `<p>An old <b>falcon</b> in the sky.</p>`
		body += fmt.Sprintf("<p>Disk exceeds the %s%% threshold. \n %s</p>", diskConfig["pct"], diskStatStr)

		request := Mail{
			Sender:  from,
			To:      to,
			Subject: subject,
			Body:    body,
		}
		// Create authentication
		auth := smtp.PlainAuth("", from, password, smtpHost)

		msg := BuildMessage(request)
		// Send actual message
		err = smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, []byte(msg))
		if err != nil {
			log.Fatal(err)
		}
	}

}
