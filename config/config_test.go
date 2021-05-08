package config

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestConfigReadString_InvalidJson(t *testing.T)  {
	cases := []string{
		"",
		"{},",
		"P{}",
		`{
   "name": "Binance",
   "enabled": true,
   "verbose": false,
   "httpTimeout": 15000000000,
   "websocketResponseCheckTimeout": 30000000,
   "websocketResponseMaxLimit": 7000000000,
   "websocketTrafficTimeout": 30000000000,
   "websocketOrderbookBufferLimit": 5,
   "baseCurrencies": "USD",
   "currencyPairs": {
    "requestFormat": {
     "uppercase": true
    },
    "configFormat": {
     "uppercase": true,
     "delimiter": "-"
    },
    "useGlobalFormat": true,
    "assetTypes": [
     "spot"
    ],
    "pairs": {
     "spot": {
      "enabled": "BTC-USDT",
      "available": "ETH-BTC,LTC-BTC,BNB-BTC,NEO-BTC,QTUM-ETH,EOS-ETH,SNT-ETH,BNT-ETH,GAS-BTC,BNB-ETH,BTC-USDT,ETH-USDT,OAX-ETH,DNT-ETH,MCO-ETH,MCO-BTC,WTC-BTC,WTC-ETH,LRC-BTC,LRC-ETH,QTUM-BTC,YOYO-BTC,OMG-BTC,OMG-ETH,ZRX-BTC,ZRX-ETH,STRAT-BTC,STRAT-ETH,SNGLS-BTC,BQX-BTC,BQX-ETH,KNC-BTC,KNC-ETH,FUN-BTC,FUN-ETH,SNM-BTC,SNM-ETH,NEO-ETH,IOTA-BTC,IOTA-ETH,LINK-BTC,LINK-ETH,XVG-BTC,XVG-ETH,MDA-BTC,MDA-ETH,MTL-BTC,MTL-ETH,EOS-BTC,SNT-BTC,ETC-ETH,ETC-BTC,MTH-BTC,MTH-ETH,ENG-BTC,ENG-ETH,DNT-BTC,ZEC-BTC,ZEC-ETH,BNT-BTC,AST-BTC,AST-ETH,DASH-BTC,DASH-ETH,OAX-BTC,BTG-BTC,BTG-ETH,EVX-BTC,EVX-ETH,REQ-BTC,REQ-ETH,VIB-BTC,VIB-ETH,TRX-BTC,TRX-ETH,POWR-BTC,POWR-ETH,ARK-BTC,ARK-ETH,YOYO-ETH,XRP-BTC,XRP-ETH,ENJ-BTC,ENJ-ETH,STORJ-BTC,STORJ-ETH,BNB-USDT,YOYO-BNB,POWR-BNB,KMD-BTC,KMD-ETH,NULS-BNB,RCN-BTC,RCN-ETH,RCN-BNB,NULS-BTC,NULS-ETH,RDN-BTC,RDN-ETH,RDN-BNB,XMR-BTC,XMR-ETH,DLT-BNB,WTC-BNB,DLT-BTC,DLT-ETH,AMB-BTC,AMB-ETH,AMB-BNB,BAT-BTC,BAT-ETH,BAT-BNB,BCPT-BTC,BCPT-ETH,BCPT-BNB,ARN-BTC,ARN-ETH,GVT-BTC,GVT-ETH,CDT-BTC,CDT-ETH,GXS-BTC,GXS-ETH,NEO-USDT,NEO-BNB,POE-BTC,POE-ETH,QSP-BTC,QSP-ETH,QSP-BNB,BTS-BTC,BTS-ETH,XZC-BTC,XZC-ETH,XZC-BNB,LSK-BTC,LSK-ETH,LSK-BNB,TNT-BTC,TNT-ETH,FUEL-BTC,MANA-BTC,MANA-ETH,BCD-BTC,BCD-ETH,DGD-BTC,DGD-ETH,IOTA-BNB,ADX-BTC,ADX-ETH,ADA-BTC,ADA-ETH,PPT-BTC,PPT-ETH,CMT-BTC,CMT-ETH,CMT-BNB,XLM-BTC,XLM-ETH,XLM-BNB,CND-BTC,CND-ETH,CND-BNB,LEND-BTC,LEND-ETH,WABI-BTC,WABI-ETH,WABI-BNB,LTC-ETH,LTC-USDT,LTC-BNB,TNB-BTC,TNB-ETH,WAVES-BTC,WAVES-ETH,WAVES-BNB,GTO-BTC,GTO-ETH,GTO-BNB,ICX-BTC,ICX-ETH,ICX-BNB,OST-BTC,OST-ETH,OST-BNB,ELF-BTC,ELF-ETH,AION-BTC,AION-ETH,AION-BNB,NEBL-BTC,NEBL-ETH,NEBL-BNB,BRD-BTC,BRD-ETH,BRD-BNB,MCO-BNB,EDO-BTC,EDO-ETH,NAV-BTC,LUN-BTC,APPC-BTC,APPC-ETH,APPC-BNB,VIBE-BTC,VIBE-ETH,RLC-BTC,RLC-ETH,RLC-BNB,INS-BTC,INS-ETH,PIVX-BTC,PIVX-ETH,PIVX-BNB,IOST-BTC,IOST-ETH,STEEM-BTC,STEEM-ETH,STEEM-BNB,NANO-BTC,NANO-ETH,NANO-BNB,VIA-BTC,VIA-ETH,VIA-BNB,BLZ-BTC,BLZ-ETH,BLZ-BNB,AE-BTC,AE-ETH,AE-BNB,NCASH-BTC,NCASH-ETH,POA-BTC,POA-ETH,ZIL-BTC,ZIL-ETH,ZIL-BNB,ONT-BTC,ONT-ETH,ONT-BNB,STORM-BTC,STORM-ETH,STORM-BNB,QTUM-BNB,QTUM-USDT,XEM-BTC,XEM-ETH,XEM-BNB,WAN-BTC,WAN-ETH,WAN-BNB,WPR-BTC,WPR-ETH,QLC-BTC,QLC-ETH,SYS-BTC,SYS-ETH,SYS-BNB,QLC-BNB,GRS-BTC,GRS-ETH,ADA-USDT,ADA-BNB,GNT-BTC,GNT-ETH,LOOM-BTC,LOOM-ETH,LOOM-BNB,XRP-USDT,REP-BTC,REP-ETH,BTC-TUSD,ETH-TUSD,ZEN-BTC,ZEN-ETH,ZEN-BNB,SKY-BTC,SKY-ETH,SKY-BNB,EOS-USDT,EOS-BNB,CVC-BTC,CVC-ETH,THETA-BTC,THETA-ETH,THETA-BNB,XRP-BNB,TUSD-USDT,IOTA-USDT,XLM-USDT,IOTX-BTC,IOTX-ETH,QKC-BTC,QKC-ETH,AGI-BTC,AGI-ETH,AGI-BNB,NXS-BTC,NXS-ETH,NXS-BNB,ENJ-BNB,DATA-BTC,DATA-ETH,ONT-USDT,TRX-BNB,TRX-USDT,ETC-USDT,ETC-BNB,ICX-USDT,SC-BTC,SC-ETH,SC-BNB,NPXS-ETH,KEY-BTC,KEY-ETH,NAS-BTC,NAS-ETH,NAS-BNB,MFT-BTC,MFT-ETH,MFT-BNB,DENT-ETH,ARDR-BTC,ARDR-ETH,NULS-USDT,HOT-BTC,HOT-ETH,VET-BTC,VET-ETH,VET-USDT,VET-BNB,DOCK-BTC,DOCK-ETH,POLY-BTC,POLY-BNB,HC-BTC,HC-ETH,GO-BTC,GO-BNB,PAX-USDT,RVN-BTC,RVN-BNB,DCR-BTC,DCR-BNB,MITH-BTC,MITH-BNB,BNB-PAX,BTC-PAX,ETH-PAX,XRP-PAX,EOS-PAX,XLM-PAX,REN-BTC,REN-BNB,BNB-TUSD,XRP-TUSD,EOS-TUSD,XLM-TUSD,BNB-USDC,BTC-USDC,ETH-USDC,XRP-USDC,EOS-USDC,XLM-USDC,USDC-USDT,ADA-TUSD,TRX-TUSD,NEO-TUSD,TRX-XRP,XZC-XRP,PAX-TUSD,USDC-TUSD,USDC-PAX,LINK-USDT,LINK-TUSD,LINK-PAX,LINK-USDC,WAVES-USDT,WAVES-TUSD,WAVES-USDC,LTC-TUSD,LTC-PAX,LTC-USDC,TRX-PAX,TRX-USDC,BTT-BNB,BTT-USDT,BNB-USDS,BTC-USDS,USDS-USDT,USDS-PAX,USDS-TUSD,USDS-USDC,BTT-PAX,BTT-TUSD,BTT-USDC,ONG-BNB,ONG-BTC,ONG-USDT,HOT-BNB,HOT-USDT,ZIL-USDT,ZRX-BNB,ZRX-USDT,FET-BNB,FET-BTC,FET-USDT,BAT-USDT,XMR-BNB,XMR-USDT,ZEC-BNB,ZEC-USDT,ZEC-PAX,ZEC-TUSD,ZEC-USDC,IOST-BNB,IOST-USDT,CELR-BNB,CELR-BTC,CELR-USDT,ADA-PAX,ADA-USDC,NEO-PAX,NEO-USDC,DASH-BNB,DASH-USDT,NANO-USDT,OMG-BNB,OMG-USDT,THETA-USDT,ENJ-USDT,MITH-USDT,MATIC-BNB,MATIC-BTC,MATIC-USDT,ATOM-BNB,ATOM-BTC,ATOM-USDT,ATOM-USDC,ATOM-TUSD,ETC-TUSD,BAT-USDC,BAT-PAX,BAT-TUSD,PHB-BNB,PHB-BTC,PHB-TUSD,TFUEL-BNB,TFUEL-BTC,TFUEL-USDT,ONE-BNB,ONE-BTC,ONE-USDT,ONE-USDC,FTM-BNB,FTM-BTC,FTM-USDT,FTM-USDC,ALGO-BNB,ALGO-BTC,ALGO-USDT,ALGO-TUSD,ALGO-PAX,ALGO-USDC,GTO-USDT,ERD-BNB,ERD-BTC,ERD-USDT,DOGE-BNB,DOGE-BTC,DOGE-USDT,DUSK-BNB,DUSK-BTC,DUSK-USDT,DUSK-USDC,DUSK-PAX,BGBP-USDC,ANKR-BNB,ANKR-BTC,ANKR-USDT,ONT-PAX,ONT-USDC,WIN-BNB,WIN-USDT,WIN-USDC,COS-BNB,COS-BTC,COS-USDT,NPXS-USDT,COCOS-BNB,COCOS-BTC,COCOS-USDT,MTL-USDT,TOMO-BNB,TOMO-BTC,TOMO-USDT,TOMO-USDC,PERL-BNB,PERL-BTC,PERL-USDT,DENT-USDT,MFT-USDT,KEY-USDT,STORM-USDT,DOCK-USDT,WAN-USDT,FUN-USDT,CVC-USDT,BTT-TRX,WIN-TRX,CHZ-BNB,CHZ-BTC,CHZ-USDT,BAND-BNB,BAND-BTC,BAND-USDT,BNB-BUSD,BTC-BUSD,BUSD-USDT,BEAM-BNB,BEAM-BTC,BEAM-USDT,XTZ-BNB,XTZ-BTC,XTZ-USDT,REN-USDT,RVN-USDT,HC-USDT,HBAR-BNB,HBAR-BTC,HBAR-USDT,NKN-BNB,NKN-BTC,NKN-USDT,XRP-BUSD,ETH-BUSD,LTC-BUSD,LINK-BUSD,ETC-BUSD,STX-BNB,STX-BTC,STX-USDT,KAVA-BNB,KAVA-BTC,KAVA-USDT,BUSD-NGN,BNB-NGN,BTC-NGN,ARPA-BNB,ARPA-BTC,ARPA-USDT,TRX-BUSD,EOS-BUSD,IOTX-USDT,RLC-USDT,MCO-USDT,XLM-BUSD,ADA-BUSD,CTXC-BNB,CTXC-BTC,CTXC-USDT,BCH-BNB,BCH-BTC,BCH-USDT,BCH-USDC,BCH-TUSD,BCH-PAX,BCH-BUSD,BTC-RUB,ETH-RUB,XRP-RUB,BNB-RUB,TROY-BNB,TROY-BTC,TROY-USDT,BUSD-RUB,QTUM-BUSD,VET-BUSD"
     }
    }
   },
   "api": {
    "authenticatedSupport": false,
    "authenticatedWebsocketApiSupport": false,
    "endpoints": {
     "url": "NON_DEFAULT_HTTP_LINK_TO_EXCHANGE_API",
     "urlSecondary": "NON_DEFAULT_HTTP_LINK_TO_EXCHANGE_API",
     "websocketURL": "NON_DEFAULT_HTTP_LINK_TO_WEBSOCKET_EXCHANGE_API"
    },
    "credentials": {
     "key": "Key",
     "secret": "Secret"
    },
    "credentialsValidator": {
     "requiresKey": true,
     "requiresSecret": true
    }
   },
   "features": {
    "supports": {
     "restAPI": true,
     "restCapabilities": {
      "tickerBatching": true,
      "autoPairUpdates": true
     },
     "websocketAPI": true,
     "websocketCapabilities": {}
    },
    "enabled": {
     "autoPairUpdates": true,
     "websocketAPI": false
    }
   },
   "bankAccounts": [
    {
     "enabled": false,
     "bankName": "",
     "bankAddress": "",
     "bankPostalCode": "",
     "bankPostalCity": "",
     "bankCountry": "",
     "accountName": "",
     "accountNumber": "",
     "swiftCode": "",
     "iban": "",
     "supportedCurrencies": ""
    }
   ]
  },
  {
   "name": "Bitfinex",
   "enabled": true,
   "verbose": false,
   "httpTimeout": 15000000000,
   "websocketResponseCheckTimeout": 30000000,
   "websocketResponseMaxLimit": 7000000000,
   "websocketTrafficTimeout": 30000000000,
   "websocketOrderbookBufferLimit": 5,
   "baseCurrencies": "USD",
   "currencyPairs": {
    "requestFormat": {
     "uppercase": true
    },
    "configFormat": {
     "uppercase": true
    },
    "useGlobalFormat": true,
    "assetTypes": [
     "spot"
    ],
    "pairs": {
     "spot": {
      "enabled": "BTCUSD,LTCUSD,LTCBTC,ETHUSD,ETHBTC",
      "available": "BTCUSD,LTCUSD,LTCBTC,ETHUSD,ETHBTC,ETCBTC,ETCUSD,RRTUSD,RRTBTC,ZECUSD,ZECBTC,XMRUSD,XMRBTC,DSHUSD,DSHBTC,BTCEUR,BTCJPY,XRPUSD,XRPBTC,IOTUSD,IOTBTC,IOTETH,EOSUSD,EOSBTC,EOSETH,SANUSD,SANBTC,SANETH,OMGUSD,OMGBTC,OMGETH,NEOUSD,NEOBTC,NEOETH,ETPUSD,ETPBTC,ETPETH,QTMUSD,QTMBTC,QTMETH,AVTUSD,AVTBTC,AVTETH,EDOUSD,EDOBTC,EDOETH,BTGUSD,BTGBTC,DATUSD,DATBTC,DATETH,QSHUSD,QSHBTC,QSHETH,YYWUSD,YYWBTC,YYWETH,GNTUSD,GNTBTC,GNTETH,SNTUSD,SNTBTC,SNTETH,IOTEUR,BATUSD,BATBTC,BATETH,MNAUSD,MNABTC,MNAETH,FUNUSD,FUNBTC,FUNETH,ZRXUSD,ZRXBTC,ZRXETH,TNBUSD,TNBBTC,TNBETH,SPKUSD,SPKBTC,SPKETH,TRXUSD,TRXBTC,TRXETH,RCNUSD,RCNBTC,RCNETH,RLCUSD,RLCBTC,RLCETH,AIDUSD,AIDBTC,AIDETH,SNGUSD,SNGBTC,SNGETH,REPUSD,REPBTC,REPETH,ELFUSD,ELFBTC,ELFETH,NECUSD,NECBTC,NECETH,BTCGBP,ETHEUR,ETHJPY,ETHGBP,NEOEUR,NEOJPY,NEOGBP,EOSEUR,EOSJPY,EOSGBP,IOTJPY,IOTGBP,IOSUSD,IOSBTC,IOSETH,AIOUSD,AIOBTC,AIOETH,REQUSD,REQBTC,REQETH,RDNUSD,RDNBTC,RDNETH,LRCUSD,LRCBTC,LRCETH,WAXUSD,WAXBTC,WAXETH,DAIUSD,DAIBTC,DAIETH,AGIUSD,AGIBTC,AGIETH,BFTUSD,BFTBTC,BFTETH,MTNUSD,MTNBTC,MTNETH,ODEUSD,ODEBTC,ODEETH,ANTUSD,ANTBTC,ANTETH,DTHUSD,DTHBTC,DTHETH,MITUSD,MITBTC,MITETH,STJUSD,STJBTC,STJETH,XLMUSD,XLMEUR,XLMJPY,XLMGBP,XLMBTC,XLMETH,XVGUSD,XVGEUR,XVGJPY,XVGGBP,XVGBTC,XVGETH,BCIUSD,BCIBTC,MKRUSD,MKRBTC,MKRETH,KNCUSD,KNCBTC,KNCETH,POAUSD,POABTC,POAETH,EVTUSD,LYMUSD,LYMBTC,LYMETH,UTKUSD,UTKBTC,UTKETH,VEEUSD,VEEBTC,VEEETH,DADUSD,DADBTC,DADETH,ORSUSD,ORSBTC,ORSETH,AUCUSD,AUCBTC,AUCETH,POYUSD,POYBTC,POYETH,FSNUSD,FSNBTC,FSNETH,CBTUSD,CBTBTC,CBTETH,ZCNUSD,ZCNBTC,ZCNETH,SENUSD,SENBTC,SENETH,NCAUSD,NCABTC,NCAETH,CNDUSD,CNDBTC,CNDETH,CTXUSD,CTXBTC,CTXETH,PAIUSD,PAIBTC,SEEUSD,SEEBTC,SEEETH,ESSUSD,ESSBTC,ESSETH,ATMUSD,ATMBTC,ATMETH,HOTUSD,HOTBTC,HOTETH,DTAUSD,DTABTC,DTAETH,IQXUSD,IQXBTC,IQXEOS,WPRUSD,WPRBTC,WPRETH,ZILUSD,ZILBTC,ZILETH,BNTUSD,BNTBTC,BNTETH,ABSUSD,ABSETH,XRAUSD,XRAETH,MANUSD,MANETH,BBNUSD,BBNETH,NIOUSD,NIOETH,DGXUSD,DGXETH,VETUSD,VETBTC,VETETH,UTNUSD,UTNETH,TKNUSD,TKNETH,GOTUSD,GOTEUR,GOTETH,XTZUSD,XTZBTC,CNNUSD,CNNETH,BOXUSD,BOXETH,TRXEUR,TRXGBP,TRXJPY,MGOUSD,MGOETH,RTEUSD,RTEETH,YGGUSD,YGGETH,MLNUSD,MLNETH,WTCUSD,WTCETH,CSXUSD,CSXETH,OMNUSD,OMNBTC,INTUSD,INTETH,DRNUSD,DRNETH,PNKUSD,PNKETH,DGBUSD,DGBBTC,BSVUSD,BSVBTC,BABUSD,BABBTC,WLOUSD,WLOXLM,VLDUSD,VLDETH,ENJUSD,ENJETH,ONLUSD,ONLETH,RBTUSD,RBTBTC,USTUSD,EUTEUR,EUTUSD,GSDUSD,UDCUSD,TSDUSD,PAXUSD,RIFUSD,RIFBTC,PASUSD,PASETH,VSYUSD,VSYBTC,ZRXDAI,MKRDAI,OMGDAI,BTTUSD,BTTBTC,BTCUST,ETHUST,CLOUSD,CLOBTC,IMPUSD,IMPETH,LTCUST,EOSUST,BABUST,SCRUSD,SCRETH,GNOUSD,GNOETH,GENUSD,GENETH,ATOUSD,ATOBTC,ATOETH,WBTUSD,XCHUSD,EUSUSD,WBTETH,XCHETH,EUSETH,LEOUSD,LEOBTC,LEOUST,LEOEOS,LEOETH,ASTUSD,ASTETH,FOAUSD,FOAETH,UFRUSD,UFRETH,ZBTUSD,ZBTUST,OKBUSD,USKUSD,GTXUSD,KANUSD,OKBUST,OKBETH,OKBBTC,USKUST,USKETH,USKBTC,USKEOS,GTXUST,KANUST,AMPUSD,ALGUSD,ALGBTC,ALGUST,BTCXCH,SWMUSD,SWMETH,TRIUSD,TRIETH,LOOUSD,LOOETH,AMPUST,DUSK:USD,DUSK:BTC,UOSUSD,UOSBTC,RRBUSD,RRBUST,DTXUSD,DTXUST,AMPBTC,FTTUSD,FTTUST,PAXUST,UDCUST,TSDUST,BTC:CNHT,UST:CNHT,CNH:CNHT,CHZUSD,CHZUST,BTCF0:USTF0,ETHF0:USTF0"
     }
    }
   },
   "api": {
    "authenticatedSupport": false,
    "authenticatedWebsocketApiSupport": false,
    "endpoints": {
     "url": "NON_DEFAULT_HTTP_LINK_TO_EXCHANGE_API",
     "urlSecondary": "NON_DEFAULT_HTTP_LINK_TO_EXCHANGE_API",
     "websocketURL": "NON_DEFAULT_HTTP_LINK_TO_WEBSOCKET_EXCHANGE_API"
    },
    "credentials": {
     "key": "Key",
     "secret": "Secret"
    },
    "credentialsValidator": {
     "requiresKey": true,
     "requiresSecret": true
    }
   },
   "features": {
    "supports": {
     "restAPI": true,
     "restCapabilities": {
      "tickerBatching": true,
      "autoPairUpdates": true
     },
     "websocketAPI": true,
     "websocketCapabilities": {}
    },
    "enabled": {
     "autoPairUpdates": true,
     "websocketAPI": false
    }
   },
   "bankAccounts": [
    {
     "enabled": false,
     "bankName": "Deutsche Bank Privat Und Geschaeftskunden AG",
     "bankAddress": "Karlsruhe, 76125, GERMANY",
     "bankPostalCode": "",
     "bankPostalCity": "",
     "bankCountry": "",
     "accountName": "GLOBAL TRADE SOLUTIONS GmbH",
     "accountNumber": "DE51660700240057016802",
     "swiftCode": "DEUTDEDB660",
     "iban": "DE51660700240057016802",
     "supportedCurrencies": "EUR,USD"
    },
    {
     "enabled": true,
     "bankName": "Deutsche Bank Privat Und Geschaeftskunden AG",
     "bankAddress": "Karlsruhe, 76125, GERMANY",
     "bankPostalCode": "",
     "bankPostalCity": "",
     "bankCountry": "",
     "accountName": "GLOBAL TRADE SOLUTIONS GmbH",
     "accountNumber": "DE78660700240057016801",
     "swiftCode": "DEUTDEDB660",
     "iban": "DE78660700240057016801",
     "supportedCurrencies": "JPY,GBP"
    }
   ]
  }`,
	}
	for _, c := range cases {
		cfg, err := FromString(c)
		assert.NotNil(t, err, c)
		assert.Nil(t, cfg, c)
	}
}

func TestConfigReadString_ValidJson(t *testing.T)  {
	validJson := `{
   "name": "Binance",
   "enabled": true,
   "verbose": false,
   "httpTimeout": 15000000000,
   "websocketResponseCheckTimeout": 30000000,
   "websocketResponseMaxLimit": 7000000000,
   "websocketTrafficTimeout": 30000000000,
   "websocketOrderbookBufferLimit": 5,
   "baseCurrencies": "USD",
   "currencyPairs": {
    "requestFormat": {
     "uppercase": true
    },
    "configFormat": {
     "uppercase": true,
     "delimiter": "-"
    },
    "useGlobalFormat": true,
    "assetTypes": [
     "spot"
    ],
    "pairs": {
     "spot": {
      "enabled": "BTC-USDT",
      "available": "ETH-BTC,LTC-BTC,BNB-BTC,NEO-BTC,QTUM-ETH,EOS-ETH,SNT-ETH,BNT-ETH,GAS-BTC,BNB-ETH,BTC-USDT,ETH-USDT,OAX-ETH,DNT-ETH,MCO-ETH,MCO-BTC,WTC-BTC,WTC-ETH,LRC-BTC,LRC-ETH,QTUM-BTC,YOYO-BTC,OMG-BTC,OMG-ETH,ZRX-BTC,ZRX-ETH,STRAT-BTC,STRAT-ETH,SNGLS-BTC,BQX-BTC,BQX-ETH,KNC-BTC,KNC-ETH,FUN-BTC,FUN-ETH,SNM-BTC,SNM-ETH,NEO-ETH,IOTA-BTC,IOTA-ETH,LINK-BTC,LINK-ETH,XVG-BTC,XVG-ETH,MDA-BTC,MDA-ETH,MTL-BTC,MTL-ETH,EOS-BTC,SNT-BTC,ETC-ETH,ETC-BTC,MTH-BTC,MTH-ETH,ENG-BTC,ENG-ETH,DNT-BTC,ZEC-BTC,ZEC-ETH,BNT-BTC,AST-BTC,AST-ETH,DASH-BTC,DASH-ETH,OAX-BTC,BTG-BTC,BTG-ETH,EVX-BTC,EVX-ETH,REQ-BTC,REQ-ETH,VIB-BTC,VIB-ETH,TRX-BTC,TRX-ETH,POWR-BTC,POWR-ETH,ARK-BTC,ARK-ETH,YOYO-ETH,XRP-BTC,XRP-ETH,ENJ-BTC,ENJ-ETH,STORJ-BTC,STORJ-ETH,BNB-USDT,YOYO-BNB,POWR-BNB,KMD-BTC,KMD-ETH,NULS-BNB,RCN-BTC,RCN-ETH,RCN-BNB,NULS-BTC,NULS-ETH,RDN-BTC,RDN-ETH,RDN-BNB,XMR-BTC,XMR-ETH,DLT-BNB,WTC-BNB,DLT-BTC,DLT-ETH,AMB-BTC,AMB-ETH,AMB-BNB,BAT-BTC,BAT-ETH,BAT-BNB,BCPT-BTC,BCPT-ETH,BCPT-BNB,ARN-BTC,ARN-ETH,GVT-BTC,GVT-ETH,CDT-BTC,CDT-ETH,GXS-BTC,GXS-ETH,NEO-USDT,NEO-BNB,POE-BTC,POE-ETH,QSP-BTC,QSP-ETH,QSP-BNB,BTS-BTC,BTS-ETH,XZC-BTC,XZC-ETH,XZC-BNB,LSK-BTC,LSK-ETH,LSK-BNB,TNT-BTC,TNT-ETH,FUEL-BTC,MANA-BTC,MANA-ETH,BCD-BTC,BCD-ETH,DGD-BTC,DGD-ETH,IOTA-BNB,ADX-BTC,ADX-ETH,ADA-BTC,ADA-ETH,PPT-BTC,PPT-ETH,CMT-BTC,CMT-ETH,CMT-BNB,XLM-BTC,XLM-ETH,XLM-BNB,CND-BTC,CND-ETH,CND-BNB,LEND-BTC,LEND-ETH,WABI-BTC,WABI-ETH,WABI-BNB,LTC-ETH,LTC-USDT,LTC-BNB,TNB-BTC,TNB-ETH,WAVES-BTC,WAVES-ETH,WAVES-BNB,GTO-BTC,GTO-ETH,GTO-BNB,ICX-BTC,ICX-ETH,ICX-BNB,OST-BTC,OST-ETH,OST-BNB,ELF-BTC,ELF-ETH,AION-BTC,AION-ETH,AION-BNB,NEBL-BTC,NEBL-ETH,NEBL-BNB,BRD-BTC,BRD-ETH,BRD-BNB,MCO-BNB,EDO-BTC,EDO-ETH,NAV-BTC,LUN-BTC,APPC-BTC,APPC-ETH,APPC-BNB,VIBE-BTC,VIBE-ETH,RLC-BTC,RLC-ETH,RLC-BNB,INS-BTC,INS-ETH,PIVX-BTC,PIVX-ETH,PIVX-BNB,IOST-BTC,IOST-ETH,STEEM-BTC,STEEM-ETH,STEEM-BNB,NANO-BTC,NANO-ETH,NANO-BNB,VIA-BTC,VIA-ETH,VIA-BNB,BLZ-BTC,BLZ-ETH,BLZ-BNB,AE-BTC,AE-ETH,AE-BNB,NCASH-BTC,NCASH-ETH,POA-BTC,POA-ETH,ZIL-BTC,ZIL-ETH,ZIL-BNB,ONT-BTC,ONT-ETH,ONT-BNB,STORM-BTC,STORM-ETH,STORM-BNB,QTUM-BNB,QTUM-USDT,XEM-BTC,XEM-ETH,XEM-BNB,WAN-BTC,WAN-ETH,WAN-BNB,WPR-BTC,WPR-ETH,QLC-BTC,QLC-ETH,SYS-BTC,SYS-ETH,SYS-BNB,QLC-BNB,GRS-BTC,GRS-ETH,ADA-USDT,ADA-BNB,GNT-BTC,GNT-ETH,LOOM-BTC,LOOM-ETH,LOOM-BNB,XRP-USDT,REP-BTC,REP-ETH,BTC-TUSD,ETH-TUSD,ZEN-BTC,ZEN-ETH,ZEN-BNB,SKY-BTC,SKY-ETH,SKY-BNB,EOS-USDT,EOS-BNB,CVC-BTC,CVC-ETH,THETA-BTC,THETA-ETH,THETA-BNB,XRP-BNB,TUSD-USDT,IOTA-USDT,XLM-USDT,IOTX-BTC,IOTX-ETH,QKC-BTC,QKC-ETH,AGI-BTC,AGI-ETH,AGI-BNB,NXS-BTC,NXS-ETH,NXS-BNB,ENJ-BNB,DATA-BTC,DATA-ETH,ONT-USDT,TRX-BNB,TRX-USDT,ETC-USDT,ETC-BNB,ICX-USDT,SC-BTC,SC-ETH,SC-BNB,NPXS-ETH,KEY-BTC,KEY-ETH,NAS-BTC,NAS-ETH,NAS-BNB,MFT-BTC,MFT-ETH,MFT-BNB,DENT-ETH,ARDR-BTC,ARDR-ETH,NULS-USDT,HOT-BTC,HOT-ETH,VET-BTC,VET-ETH,VET-USDT,VET-BNB,DOCK-BTC,DOCK-ETH,POLY-BTC,POLY-BNB,HC-BTC,HC-ETH,GO-BTC,GO-BNB,PAX-USDT,RVN-BTC,RVN-BNB,DCR-BTC,DCR-BNB,MITH-BTC,MITH-BNB,BNB-PAX,BTC-PAX,ETH-PAX,XRP-PAX,EOS-PAX,XLM-PAX,REN-BTC,REN-BNB,BNB-TUSD,XRP-TUSD,EOS-TUSD,XLM-TUSD,BNB-USDC,BTC-USDC,ETH-USDC,XRP-USDC,EOS-USDC,XLM-USDC,USDC-USDT,ADA-TUSD,TRX-TUSD,NEO-TUSD,TRX-XRP,XZC-XRP,PAX-TUSD,USDC-TUSD,USDC-PAX,LINK-USDT,LINK-TUSD,LINK-PAX,LINK-USDC,WAVES-USDT,WAVES-TUSD,WAVES-USDC,LTC-TUSD,LTC-PAX,LTC-USDC,TRX-PAX,TRX-USDC,BTT-BNB,BTT-USDT,BNB-USDS,BTC-USDS,USDS-USDT,USDS-PAX,USDS-TUSD,USDS-USDC,BTT-PAX,BTT-TUSD,BTT-USDC,ONG-BNB,ONG-BTC,ONG-USDT,HOT-BNB,HOT-USDT,ZIL-USDT,ZRX-BNB,ZRX-USDT,FET-BNB,FET-BTC,FET-USDT,BAT-USDT,XMR-BNB,XMR-USDT,ZEC-BNB,ZEC-USDT,ZEC-PAX,ZEC-TUSD,ZEC-USDC,IOST-BNB,IOST-USDT,CELR-BNB,CELR-BTC,CELR-USDT,ADA-PAX,ADA-USDC,NEO-PAX,NEO-USDC,DASH-BNB,DASH-USDT,NANO-USDT,OMG-BNB,OMG-USDT,THETA-USDT,ENJ-USDT,MITH-USDT,MATIC-BNB,MATIC-BTC,MATIC-USDT,ATOM-BNB,ATOM-BTC,ATOM-USDT,ATOM-USDC,ATOM-TUSD,ETC-TUSD,BAT-USDC,BAT-PAX,BAT-TUSD,PHB-BNB,PHB-BTC,PHB-TUSD,TFUEL-BNB,TFUEL-BTC,TFUEL-USDT,ONE-BNB,ONE-BTC,ONE-USDT,ONE-USDC,FTM-BNB,FTM-BTC,FTM-USDT,FTM-USDC,ALGO-BNB,ALGO-BTC,ALGO-USDT,ALGO-TUSD,ALGO-PAX,ALGO-USDC,GTO-USDT,ERD-BNB,ERD-BTC,ERD-USDT,DOGE-BNB,DOGE-BTC,DOGE-USDT,DUSK-BNB,DUSK-BTC,DUSK-USDT,DUSK-USDC,DUSK-PAX,BGBP-USDC,ANKR-BNB,ANKR-BTC,ANKR-USDT,ONT-PAX,ONT-USDC,WIN-BNB,WIN-USDT,WIN-USDC,COS-BNB,COS-BTC,COS-USDT,NPXS-USDT,COCOS-BNB,COCOS-BTC,COCOS-USDT,MTL-USDT,TOMO-BNB,TOMO-BTC,TOMO-USDT,TOMO-USDC,PERL-BNB,PERL-BTC,PERL-USDT,DENT-USDT,MFT-USDT,KEY-USDT,STORM-USDT,DOCK-USDT,WAN-USDT,FUN-USDT,CVC-USDT,BTT-TRX,WIN-TRX,CHZ-BNB,CHZ-BTC,CHZ-USDT,BAND-BNB,BAND-BTC,BAND-USDT,BNB-BUSD,BTC-BUSD,BUSD-USDT,BEAM-BNB,BEAM-BTC,BEAM-USDT,XTZ-BNB,XTZ-BTC,XTZ-USDT,REN-USDT,RVN-USDT,HC-USDT,HBAR-BNB,HBAR-BTC,HBAR-USDT,NKN-BNB,NKN-BTC,NKN-USDT,XRP-BUSD,ETH-BUSD,LTC-BUSD,LINK-BUSD,ETC-BUSD,STX-BNB,STX-BTC,STX-USDT,KAVA-BNB,KAVA-BTC,KAVA-USDT,BUSD-NGN,BNB-NGN,BTC-NGN,ARPA-BNB,ARPA-BTC,ARPA-USDT,TRX-BUSD,EOS-BUSD,IOTX-USDT,RLC-USDT,MCO-USDT,XLM-BUSD,ADA-BUSD,CTXC-BNB,CTXC-BTC,CTXC-USDT,BCH-BNB,BCH-BTC,BCH-USDT,BCH-USDC,BCH-TUSD,BCH-PAX,BCH-BUSD,BTC-RUB,ETH-RUB,XRP-RUB,BNB-RUB,TROY-BNB,TROY-BTC,TROY-USDT,BUSD-RUB,QTUM-BUSD,VET-BUSD"
     }
    }
   },
   "api": {
    "authenticatedSupport": false,
    "authenticatedWebsocketApiSupport": false,
    "endpoints": {
     "url": "NON_DEFAULT_HTTP_LINK_TO_EXCHANGE_API",
     "urlSecondary": "NON_DEFAULT_HTTP_LINK_TO_EXCHANGE_API",
     "websocketURL": "NON_DEFAULT_HTTP_LINK_TO_WEBSOCKET_EXCHANGE_API"
    },
    "credentials": {
     "key": "Key",
     "secret": "Secret"
    },
    "credentialsValidator": {
     "requiresKey": true,
     "requiresSecret": true
    }
   },
   "features": {
    "supports": {
     "restAPI": true,
     "restCapabilities": {
      "tickerBatching": true,
      "autoPairUpdates": true
     },
     "websocketAPI": true,
     "websocketCapabilities": {}
    },
    "enabled": {
     "autoPairUpdates": true,
     "websocketAPI": false
    }
   },
   "bankAccounts": [
    {
     "enabled": false,
     "bankName": "",
     "bankAddress": "",
     "bankPostalCode": "",
     "bankPostalCity": "",
     "bankCountry": "",
     "accountName": "",
     "accountNumber": "",
     "swiftCode": "",
     "iban": "",
     "supportedCurrencies": ""
    }
   ]
  }
  `
	cfg, err := FromString(validJson)
	if err != nil || cfg == nil {
		t.Fail()
	}
}

func TestFromReader(t *testing.T) {
	testCases := []struct{
		data []byte
		shouldError bool
	}{
		{[]byte("{},"), true},
		{[]byte(nil), true},
		{[]byte("p{}"), true},
		{[]byte(`{
   "name": "Binance",
   "enabled": true,
   "verbose": false,
   "httpTimeout": 15000000000,
   "websocketResponseCheckTimeout": 30000000,
   "websocketResponseMaxLimit": 7000000000,
   "websocketTrafficTimeout": 30000000000,
   "websocketOrderbookBufferLimit": 5,
   "baseCurrencies": "USD",
   "currencyPairs": {
    "requestFormat": {
     "uppercase": true
    },
    "configFormat": {
     "uppercase": true,
     "delimiter": "-"
    },
    "useGlobalFormat": true,
    "assetTypes": [
     "spot"
    ],
    "pairs": {
     "spot": {
      "enabled": "BTC-USDT",
      "available": "ETH-BTC,LTC-BTC,BNB-BTC,NEO-BTC,QTUM-ETH,EOS-ETH,SNT-ETH,BNT-ETH,GAS-BTC,BNB-ETH,BTC-USDT,ETH-USDT,OAX-ETH,DNT-ETH,MCO-ETH,MCO-BTC,WTC-BTC,WTC-ETH,LRC-BTC,LRC-ETH,QTUM-BTC,YOYO-BTC,OMG-BTC,OMG-ETH,ZRX-BTC,ZRX-ETH,STRAT-BTC,STRAT-ETH,SNGLS-BTC,BQX-BTC,BQX-ETH,KNC-BTC,KNC-ETH,FUN-BTC,FUN-ETH,SNM-BTC,SNM-ETH,NEO-ETH,IOTA-BTC,IOTA-ETH,LINK-BTC,LINK-ETH,XVG-BTC,XVG-ETH,MDA-BTC,MDA-ETH,MTL-BTC,MTL-ETH,EOS-BTC,SNT-BTC,ETC-ETH,ETC-BTC,MTH-BTC,MTH-ETH,ENG-BTC,ENG-ETH,DNT-BTC,ZEC-BTC,ZEC-ETH,BNT-BTC,AST-BTC,AST-ETH,DASH-BTC,DASH-ETH,OAX-BTC,BTG-BTC,BTG-ETH,EVX-BTC,EVX-ETH,REQ-BTC,REQ-ETH,VIB-BTC,VIB-ETH,TRX-BTC,TRX-ETH,POWR-BTC,POWR-ETH,ARK-BTC,ARK-ETH,YOYO-ETH,XRP-BTC,XRP-ETH,ENJ-BTC,ENJ-ETH,STORJ-BTC,STORJ-ETH,BNB-USDT,YOYO-BNB,POWR-BNB,KMD-BTC,KMD-ETH,NULS-BNB,RCN-BTC,RCN-ETH,RCN-BNB,NULS-BTC,NULS-ETH,RDN-BTC,RDN-ETH,RDN-BNB,XMR-BTC,XMR-ETH,DLT-BNB,WTC-BNB,DLT-BTC,DLT-ETH,AMB-BTC,AMB-ETH,AMB-BNB,BAT-BTC,BAT-ETH,BAT-BNB,BCPT-BTC,BCPT-ETH,BCPT-BNB,ARN-BTC,ARN-ETH,GVT-BTC,GVT-ETH,CDT-BTC,CDT-ETH,GXS-BTC,GXS-ETH,NEO-USDT,NEO-BNB,POE-BTC,POE-ETH,QSP-BTC,QSP-ETH,QSP-BNB,BTS-BTC,BTS-ETH,XZC-BTC,XZC-ETH,XZC-BNB,LSK-BTC,LSK-ETH,LSK-BNB,TNT-BTC,TNT-ETH,FUEL-BTC,MANA-BTC,MANA-ETH,BCD-BTC,BCD-ETH,DGD-BTC,DGD-ETH,IOTA-BNB,ADX-BTC,ADX-ETH,ADA-BTC,ADA-ETH,PPT-BTC,PPT-ETH,CMT-BTC,CMT-ETH,CMT-BNB,XLM-BTC,XLM-ETH,XLM-BNB,CND-BTC,CND-ETH,CND-BNB,LEND-BTC,LEND-ETH,WABI-BTC,WABI-ETH,WABI-BNB,LTC-ETH,LTC-USDT,LTC-BNB,TNB-BTC,TNB-ETH,WAVES-BTC,WAVES-ETH,WAVES-BNB,GTO-BTC,GTO-ETH,GTO-BNB,ICX-BTC,ICX-ETH,ICX-BNB,OST-BTC,OST-ETH,OST-BNB,ELF-BTC,ELF-ETH,AION-BTC,AION-ETH,AION-BNB,NEBL-BTC,NEBL-ETH,NEBL-BNB,BRD-BTC,BRD-ETH,BRD-BNB,MCO-BNB,EDO-BTC,EDO-ETH,NAV-BTC,LUN-BTC,APPC-BTC,APPC-ETH,APPC-BNB,VIBE-BTC,VIBE-ETH,RLC-BTC,RLC-ETH,RLC-BNB,INS-BTC,INS-ETH,PIVX-BTC,PIVX-ETH,PIVX-BNB,IOST-BTC,IOST-ETH,STEEM-BTC,STEEM-ETH,STEEM-BNB,NANO-BTC,NANO-ETH,NANO-BNB,VIA-BTC,VIA-ETH,VIA-BNB,BLZ-BTC,BLZ-ETH,BLZ-BNB,AE-BTC,AE-ETH,AE-BNB,NCASH-BTC,NCASH-ETH,POA-BTC,POA-ETH,ZIL-BTC,ZIL-ETH,ZIL-BNB,ONT-BTC,ONT-ETH,ONT-BNB,STORM-BTC,STORM-ETH,STORM-BNB,QTUM-BNB,QTUM-USDT,XEM-BTC,XEM-ETH,XEM-BNB,WAN-BTC,WAN-ETH,WAN-BNB,WPR-BTC,WPR-ETH,QLC-BTC,QLC-ETH,SYS-BTC,SYS-ETH,SYS-BNB,QLC-BNB,GRS-BTC,GRS-ETH,ADA-USDT,ADA-BNB,GNT-BTC,GNT-ETH,LOOM-BTC,LOOM-ETH,LOOM-BNB,XRP-USDT,REP-BTC,REP-ETH,BTC-TUSD,ETH-TUSD,ZEN-BTC,ZEN-ETH,ZEN-BNB,SKY-BTC,SKY-ETH,SKY-BNB,EOS-USDT,EOS-BNB,CVC-BTC,CVC-ETH,THETA-BTC,THETA-ETH,THETA-BNB,XRP-BNB,TUSD-USDT,IOTA-USDT,XLM-USDT,IOTX-BTC,IOTX-ETH,QKC-BTC,QKC-ETH,AGI-BTC,AGI-ETH,AGI-BNB,NXS-BTC,NXS-ETH,NXS-BNB,ENJ-BNB,DATA-BTC,DATA-ETH,ONT-USDT,TRX-BNB,TRX-USDT,ETC-USDT,ETC-BNB,ICX-USDT,SC-BTC,SC-ETH,SC-BNB,NPXS-ETH,KEY-BTC,KEY-ETH,NAS-BTC,NAS-ETH,NAS-BNB,MFT-BTC,MFT-ETH,MFT-BNB,DENT-ETH,ARDR-BTC,ARDR-ETH,NULS-USDT,HOT-BTC,HOT-ETH,VET-BTC,VET-ETH,VET-USDT,VET-BNB,DOCK-BTC,DOCK-ETH,POLY-BTC,POLY-BNB,HC-BTC,HC-ETH,GO-BTC,GO-BNB,PAX-USDT,RVN-BTC,RVN-BNB,DCR-BTC,DCR-BNB,MITH-BTC,MITH-BNB,BNB-PAX,BTC-PAX,ETH-PAX,XRP-PAX,EOS-PAX,XLM-PAX,REN-BTC,REN-BNB,BNB-TUSD,XRP-TUSD,EOS-TUSD,XLM-TUSD,BNB-USDC,BTC-USDC,ETH-USDC,XRP-USDC,EOS-USDC,XLM-USDC,USDC-USDT,ADA-TUSD,TRX-TUSD,NEO-TUSD,TRX-XRP,XZC-XRP,PAX-TUSD,USDC-TUSD,USDC-PAX,LINK-USDT,LINK-TUSD,LINK-PAX,LINK-USDC,WAVES-USDT,WAVES-TUSD,WAVES-USDC,LTC-TUSD,LTC-PAX,LTC-USDC,TRX-PAX,TRX-USDC,BTT-BNB,BTT-USDT,BNB-USDS,BTC-USDS,USDS-USDT,USDS-PAX,USDS-TUSD,USDS-USDC,BTT-PAX,BTT-TUSD,BTT-USDC,ONG-BNB,ONG-BTC,ONG-USDT,HOT-BNB,HOT-USDT,ZIL-USDT,ZRX-BNB,ZRX-USDT,FET-BNB,FET-BTC,FET-USDT,BAT-USDT,XMR-BNB,XMR-USDT,ZEC-BNB,ZEC-USDT,ZEC-PAX,ZEC-TUSD,ZEC-USDC,IOST-BNB,IOST-USDT,CELR-BNB,CELR-BTC,CELR-USDT,ADA-PAX,ADA-USDC,NEO-PAX,NEO-USDC,DASH-BNB,DASH-USDT,NANO-USDT,OMG-BNB,OMG-USDT,THETA-USDT,ENJ-USDT,MITH-USDT,MATIC-BNB,MATIC-BTC,MATIC-USDT,ATOM-BNB,ATOM-BTC,ATOM-USDT,ATOM-USDC,ATOM-TUSD,ETC-TUSD,BAT-USDC,BAT-PAX,BAT-TUSD,PHB-BNB,PHB-BTC,PHB-TUSD,TFUEL-BNB,TFUEL-BTC,TFUEL-USDT,ONE-BNB,ONE-BTC,ONE-USDT,ONE-USDC,FTM-BNB,FTM-BTC,FTM-USDT,FTM-USDC,ALGO-BNB,ALGO-BTC,ALGO-USDT,ALGO-TUSD,ALGO-PAX,ALGO-USDC,GTO-USDT,ERD-BNB,ERD-BTC,ERD-USDT,DOGE-BNB,DOGE-BTC,DOGE-USDT,DUSK-BNB,DUSK-BTC,DUSK-USDT,DUSK-USDC,DUSK-PAX,BGBP-USDC,ANKR-BNB,ANKR-BTC,ANKR-USDT,ONT-PAX,ONT-USDC,WIN-BNB,WIN-USDT,WIN-USDC,COS-BNB,COS-BTC,COS-USDT,NPXS-USDT,COCOS-BNB,COCOS-BTC,COCOS-USDT,MTL-USDT,TOMO-BNB,TOMO-BTC,TOMO-USDT,TOMO-USDC,PERL-BNB,PERL-BTC,PERL-USDT,DENT-USDT,MFT-USDT,KEY-USDT,STORM-USDT,DOCK-USDT,WAN-USDT,FUN-USDT,CVC-USDT,BTT-TRX,WIN-TRX,CHZ-BNB,CHZ-BTC,CHZ-USDT,BAND-BNB,BAND-BTC,BAND-USDT,BNB-BUSD,BTC-BUSD,BUSD-USDT,BEAM-BNB,BEAM-BTC,BEAM-USDT,XTZ-BNB,XTZ-BTC,XTZ-USDT,REN-USDT,RVN-USDT,HC-USDT,HBAR-BNB,HBAR-BTC,HBAR-USDT,NKN-BNB,NKN-BTC,NKN-USDT,XRP-BUSD,ETH-BUSD,LTC-BUSD,LINK-BUSD,ETC-BUSD,STX-BNB,STX-BTC,STX-USDT,KAVA-BNB,KAVA-BTC,KAVA-USDT,BUSD-NGN,BNB-NGN,BTC-NGN,ARPA-BNB,ARPA-BTC,ARPA-USDT,TRX-BUSD,EOS-BUSD,IOTX-USDT,RLC-USDT,MCO-USDT,XLM-BUSD,ADA-BUSD,CTXC-BNB,CTXC-BTC,CTXC-USDT,BCH-BNB,BCH-BTC,BCH-USDT,BCH-USDC,BCH-TUSD,BCH-PAX,BCH-BUSD,BTC-RUB,ETH-RUB,XRP-RUB,BNB-RUB,TROY-BNB,TROY-BTC,TROY-USDT,BUSD-RUB,QTUM-BUSD,VET-BUSD"
     }
    }
   },
   "api": {
    "authenticatedSupport": false,
    "authenticatedWebsocketApiSupport": false,
    "endpoints": {
     "url": "NON_DEFAULT_HTTP_LINK_TO_EXCHANGE_API",
     "urlSecondary": "NON_DEFAULT_HTTP_LINK_TO_EXCHANGE_API",
     "websocketURL": "NON_DEFAULT_HTTP_LINK_TO_WEBSOCKET_EXCHANGE_API"
    },
    "credentials": {
     "key": "Key",
     "secret": "Secret"
    },
    "credentialsValidator": {
     "requiresKey": true,
     "requiresSecret": true
    }
   },
   "features": {
    "supports": {
     "restAPI": true,
     "restCapabilities": {
      "tickerBatching": true,
      "autoPairUpdates": true
     },
     "websocketAPI": true,
     "websocketCapabilities": {}
    },
    "enabled": {
     "autoPairUpdates": true,
     "websocketAPI": false
    }
   },
   "bankAccounts": [
    {
     "enabled": false,
     "bankName": "",
     "bankAddress": "",
     "bankPostalCode": "",
     "bankPostalCity": "",
     "bankCountry": "",
     "accountName": "",
     "accountNumber": "",
     "swiftCode": "",
     "iban": "",
     "supportedCurrencies": ""
    }
   ]
  }`), false},
	}
	for _, c := range testCases {
		reader := ioutil.NopCloser(bytes.NewReader(c.data))
		cfg, err := FromReader(reader)
		if c.shouldError {
			assert.Nil(t, cfg, c)
			assert.NotNil(t, err, c)
		} else {
			assert.NotNil(t, cfg, c)
			assert.Nil(t, err, c)
		}
	}
}
func TestFromFile(t *testing.T) {
	wd, _ := os.Getwd()
	testCases := []struct{
		data string
		shouldError bool
	}{
		{"non-existen.json", true},
		{"non-existen.json", true},
		{fmt.Sprintf("%s/%s", wd, "non-existent.json"), true},
		{fmt.Sprintf("%s/%s", wd, "invalid.conf.json"), true},
		{fmt.Sprintf("%s/%s", wd, "sample.conf.json"), false},
	}
	for _, c := range testCases {
		cfg, err := FromFile(c.data)
		if c.shouldError {
			assert.Nil(t, cfg, c)
			assert.NotNil(t, err, c)
		} else {
			assert.NotNil(t, cfg, c)
			assert.Nil(t, err, c)
		}
	}
}
