package api

type ErrResponseMsg struct {
	Message string `json:"message"`
}

type HealthCheckResposne struct {
	WalletsConn     string `json:"wallets_conn"`
	WalletTransConn string `json:"wallet_trans_conn"`
	BrokerConn      string `json:"broker_conn"`
}
