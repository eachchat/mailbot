package matrix

import (
	"fmt"

	"github.com/rs/zerolog"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"
)

var LOG *zerolog.Logger

// LoginMatrix  login into matrix homeserver
func (mx *MxConf) LoginMatrix() {
	LOG.Info().Msg(fmt.Sprintf("Logging into %s as %s", mx.MatrixServer, mx.MatrixUserID))
	if len(mx.Matrixaccesstoken) > 0 {
		client, err := mautrix.NewClient(mx.MatrixServer, id.UserID(mx.MatrixUserID), mx.Matrixaccesstoken)
		if err != nil {
			panic(err)
		}
		mx.Client = client
	} else {
		client, err := mautrix.NewClient(mx.MatrixServer, "", "")
		if err != nil {
			panic(err)
		}

		res, err := client.Login(&mautrix.ReqLogin{
			Type:             "m.login.password",
			Identifier:       mautrix.UserIdentifier{Type: mautrix.IdentifierTypeUser, User: mx.MatrixUserID},
			Password:         mx.MatrixUserPassword,
			StoreCredentials: true,
		})
		if err != nil {
			panic(err)
		}
		mx.Client = client
		mx.Matrixaccesstoken = res.AccessToken
	}
	LOG.Info().Msg("Login sucessfully!")
	//mx.Client.Store = newFileStore(mx.DataDir)
	go mx.startMatrixSync()
}
