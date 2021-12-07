package fabric

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	conf "github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

func InitGateway() *gateway.Gateway {
	err := os.Setenv("DISCOVERY_AS_LOCALHOST", "true")
	if err != nil {
		log.Fatalf("Error setting DISCOVERY_AS_LOCALHOST environemnt variable: %v", err)
	}

	cfg := getConfig()
	wallet := getIdentity()

	gw, err := gateway.Connect(
		gateway.WithConfig(cfg),
		gateway.WithIdentity(wallet, "appAccount"),
	)
	// fmt.Println(gw)

	if err != nil {
		log.Fatalf("Failed to connect to gateway: %v", err)
	}
	defer gw.Close()
	return gw
}

func getIdentity() *gateway.Wallet {
	wallet, err := gateway.NewFileSystemWallet("wallet")
	if err != nil {
		log.Fatalf("Failed to create wallet: %v", err)
	}
	// fmt.Println("--------------------------")
	// fmt.Println(wallet)
	// fmt.Println("--------------------------")

	if !wallet.Exists("appAccount") {
		err = populateWallet(wallet)
		if err != nil {
			log.Fatalf("Failed to populate wallet contents: %v", err)
		}
	}
	return wallet
}

func getConfig() core.ConfigProvider {
	ccpPath1 := filepath.Join(
		"..",
		"Desktop",
		"FabricSamples",
		"fabric-samples",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"connection-org1.yaml",
	)
	// ccpPath2 := filepath.Join(
	// 	"..",
	// 	"..",
	// 	"test-network",
	// 	"quandm",
	// 	"config.yaml",
	// )
	// switch opts {
	// case "gateway":
	// 	return conf.FromFile(filepath.Clean(ccpPath1))
	// case "client":
	// 	return conf.FromFile(filepath.Clean(ccpPath2))
	// }
	return conf.FromFile(filepath.Clean(ccpPath1))
}

func populateWallet(wallet *gateway.Wallet) error {
	log.Println("============ Populating wallet ============")
	credPath := filepath.Join(
		"..",
		"Desktop",
		"FabricSamples",
		"fabric-samples",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"users",
		"User1@org1.example.com",
		"msp",
	)

	certPath := filepath.Join(credPath, "signcerts", "User1@org1.example.com-cert.pem")
	// read the certificate pem
	cert, err := ioutil.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return err
	}

	// fmt.Println("--------------------------")
	// fmt.Println("CertPath: ", certPath)
	// fmt.Println("Cert :", cert)
	// fmt.Println("--------------------------")

	keyDir := filepath.Join(credPath, "keystore")
	// there's a single file in this dir containing the private key
	files, err := ioutil.ReadDir(keyDir)
	if err != nil {
		return err
	}
	if len(files) != 1 {
		return fmt.Errorf("keystore folder should have contain one file")
	}
	keyPath := filepath.Join(keyDir, files[0].Name())
	key, err := ioutil.ReadFile(filepath.Clean(keyPath))
	if err != nil {
		return err
	}

	identity := gateway.NewX509Identity("Org1MSP", string(cert), string(key))

	return wallet.Put("appAccount", identity)
}
