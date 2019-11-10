package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/bitfield/script.v0"
)

var ()

func init() {
	log.SetFlags(0)

	mustSetEnv("DB_IMAGE", "postgres:11-alpine")
	mustSetEnv("DB_CONTAINER", "pg_host")
	mustSetEnv("DB_USER", "postgres")
	mustSetEnv("DB_PASSWD", "123")

	mustSetEnv("FLYWAY_IMAGE", "flyway/flyway:6.0.7-alpine")
	mustSetEnv("FLYWAY_URL", expand("jdbc:postgresql://$DB_CONTAINER/postgres"))
}

func expand(s string) string {
	s = os.ExpandEnv(s)
	return s
}

func mustSetEnv(name string, value string) {
	must(os.Setenv(name, value))
}

func dieIf(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}

func must(err error) {
	dieIf(err)
}

func onErrorLogToStderr(p *script.Pipe) *script.Pipe {
	if p.Error() != nil {
		io.Copy(os.Stderr, p.Reader)
	}
	return p
}

func stderr(p *script.Pipe) (n int64, err error) {
	n, err = io.Copy(os.Stderr, p.Reader)
	if p.Error() != nil {
		err = p.Error()
	}
	return
}

func waitDBIsUp(maxTimeout time.Duration) {
	waitTimeout := maxTimeout / 10
	until := time.Now().Add(maxTimeout)

	for time.Now().Before(until) {
		time.Sleep(waitTimeout)
		n, _ := script.Exec("docker ps").Match(expand("$DB_CONTAINER")).CountLines()
		if n > 0 {
			break
		}
	}
}

func exec(shellExpr string) *script.Pipe {
	s := expand(shellExpr)
	fmt.Fprintln(os.Stderr, s)
	return script.Exec(s)
}

func stopDBContainer() {
	stderr(exec("docker stop $DB_CONTAINER"))
}

func render() (err error) {
	if _, err = stderr(exec(
		"docker run --rm --name $DB_CONTAINER -e POSTGRES_PASSWORD=$DB_PASSWD -d $DB_IMAGE")); err != nil {

		return
	}
	defer stopDBContainer()

	waitDBIsUp(2 * time.Second)

	if _, err = stderr(exec(
		"docker run --rm --link $DB_CONTAINER -v $FLYWAY_PATH:/flyway/sql -t $FLYWAY_IMAGE -url=$FLYWAY_URL -user=$DB_USER -password=$DB_PASSWD migrate")); err != nil {

		return
	}

	_, err = onErrorLogToStderr(exec(
		"docker run --rm --link $DB_CONTAINER -e PGPASSWORD=$DB_PASSWD -t $DB_IMAGE pg_dump -s -h $DB_CONTAINER -U $DB_USER")).Stdout()

	return
}

func main() {
	flag.Parse()
	srcPath := flag.Arg(0)
	if srcPath == "" {
		srcPath = "."
	}

	flywayPath, err := filepath.Abs(srcPath)
	dieIf(err)

	mustSetEnv("FLYWAY_PATH", flywayPath)

	dieIf(script.IfExists(flywayPath).Error())

	must(render())
}
