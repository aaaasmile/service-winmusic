package util

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/kardianos/osext"
)

var rootPath string
var UseRelativeRoot = true

const RsaLen = 2048

func GetEncryptedPasswordFromPrompt() string {
	_, pub, err := GetKeys("cert/key.pem", true)
	if err != nil {
		log.Fatal(err)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Please enter the secret to encode")
	response, err := reader.ReadString('\n')
	response = strings.Trim(response, "\n")
	response = strings.Trim(response, "\r")
	if err != nil {
		log.Fatal(err)
	}
	buf := []byte(response)
	enc := Encrypt(buf, pub)

	return base64.StdEncoding.EncodeToString(enc)
}

func DecodePassword(secretBase64 string, privateKeyFile string, createNew bool) (string, error) {
	plain, err := base64.StdEncoding.DecodeString(secretBase64)
	if err != nil {
		return "", err
	}
	priv, _, err := GetKeys(privateKeyFile, createNew)
	if err != nil {
		log.Println("[WARNING] something went wrong with keys")
		return "", err
	}

	raw, err := Decrypt(plain, priv)
	if err != nil {
		return "", err
	}
	log.Println("Decrypting succeded")
	return string(raw), nil
}

func GetUserLogFile(serviceName string) string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	fn := filepath.Join(usr.HomeDir, "Appdata", "Local", "Temp", fmt.Sprintf("%s.log", serviceName))
	log.Println("Logfile in use is ", fn)
	return fn
}

func GetFullPath(relPath string) string {
	if UseRelativeRoot {
		return relPath
	}

	if rootPath == "" {
		var err error
		rootPath, err = osext.ExecutableFolder()
		if err != nil {
			log.Fatalf("ExecutableFolder failed: %v", err)
		}
		log.Println("Executable folder (rootdir) is ", rootPath)
		//rootPath, _ = filepath.Split(os.Args[0]) // os.Args[0] can be "faked". (https://github.com/kardianos/osext)
	}
	r := filepath.Join(rootPath, relPath)
	return r
}

func GetFullPathAbs(relPath string) string {
	if rootPath == "" {
		var err error
		rootPath, err = osext.ExecutableFolder()
		if err != nil {
			log.Fatalf("ExecutableFolder failed: %v", err)
		}
		log.Println("Executable folder (rootdir) is ", rootPath)
		//rootPath, _ = filepath.Split(os.Args[0]) // os.Args[0] can be "faked". (https://github.com/kardianos/osext)
	}
	r := filepath.Join(rootPath, relPath)
	return r
}

func GenerateGUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	guid := fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return guid
}

func GenerateTimeTrail() string {
	tt := time.Now()
	return tt.Format("20060102-150405")
}

func GetKeys(privateKeyFile string, createNew bool) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	keyFile := privateKeyFile
	mysalt := "PortalTest"
	priv, err := privateKeyFromFile(keyFile, mysalt)
	if err != nil {
		if createNew {
			log.Println("Key pem file error, create a new one", err)
			priv, _ = rsa.GenerateKey(rand.Reader, RsaLen)
			err = savePrivateKeyInFile(keyFile, priv, mysalt)
			if err != nil {
				return nil, nil, err
			}
		} else {
			return nil, nil, err
		}
	}

	return priv, &priv.PublicKey, err
}

func Encrypt(plain []byte, pubkey *rsa.PublicKey) []byte {

	//è interessante notare la procedura ibrida della criptazione.
	// Viene generata una nuova chiave random la quale viene poi criptata con la chiave pubblica
	// e messa in testa al file. La chiave della sessione viene criptata con rsa.
	// Mentre il file viene creiptato con aes che è una procedura di cifrazione simmetrica.
	key := make([]byte, 256/8) // AES-256
	io.ReadFull(rand.Reader, key)

	encKey, _ := rsa.EncryptOAEP(sha256.New(), rand.Reader, pubkey, key, nil)
	block, _ := aes.NewCipher(key)
	aesgcm, _ := cipher.NewGCM(block)
	nonce := make([]byte, aesgcm.NonceSize())
	io.ReadFull(rand.Reader, nonce)
	ciph := aesgcm.Seal(nil, nonce, plain, nil)
	s := [][]byte{encKey, nonce, ciph}
	return bytes.Join(s, []byte{})
}

func Decrypt(ciph []byte, priv *rsa.PrivateKey) ([]byte, error) {
	//Per primo viene estratta la chiave per la decriptazione via aes.
	// La chiave è in testa al file ed è codificata in rsa. La decriptazione della chiave per
	// la sessione aes è possibile solo via rsa utilizzando la chiave privata in formato pem.
	encKey := ciph[:RsaLen/8]
	ciph = ciph[RsaLen/8:]
	key, _ := rsa.DecryptOAEP(sha256.New(), rand.Reader, priv, encKey, nil)

	block, _ := aes.NewCipher(key)
	aesgcm, _ := cipher.NewGCM(block)
	nonce := ciph[:aesgcm.NonceSize()]
	ciph = ciph[aesgcm.NonceSize():]

	return aesgcm.Open(nil, nonce, ciph, nil)
}

func savePrivateKeyInFile(file string, priv *rsa.PrivateKey, pwd string) error {
	der := x509.MarshalPKCS1PrivateKey(priv)
	pp := []byte(pwd)
	block, err := x509.EncryptPEMBlock(rand.Reader, "RSA PRIVATE KEY", der, pp, x509.PEMCipherAES256)
	if err != nil {
		return err
	}
	log.Println("Save the key in ", file)
	return ioutil.WriteFile(file, pem.EncodeToMemory(block), 0644)
}

func privateKeyFromFile(file string, pwd string) (*rsa.PrivateKey, error) {
	der, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(der)

	der, err = x509.DecryptPEMBlock(block, []byte(pwd))
	if err != nil {
		return nil, err
	}
	priv, err := x509.ParsePKCS1PrivateKey(der)
	return priv, nil
}
