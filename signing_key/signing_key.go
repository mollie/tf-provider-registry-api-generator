package signing_key

import (
	"io/ioutil"
	"log"
	"os/exec"
)

type PGPSigningKey struct {
	KeyID      string
	ASCIIArmor string
}

func GetPublicSigningKey(fingerPrint string) PGPSigningKey {


	cmd := exec.Command("gpg", "--armor", "--export", fingerPrint)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("%s", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatalf("%s", err)
	}

	err = cmd.Start()
	if err != nil {
		log.Fatalf("%s", err)
	}
	key, err := ioutil.ReadAll(stdout)
	if err != nil {
		log.Fatalf("%s", err)
	}
	if len(key) == 0 {
		msg,_ := ioutil.ReadAll(stderr)
		log.Fatalf("ERROR: failed to retrieve public key %s, %s", fingerPrint, string(msg))
	}
	return PGPSigningKey{fingerPrint, string(key)}
}
