package examples

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
)

func init() {
	setupFuncs["SagaBarrierSetup"] = func(app *gin.Engine) {
		app.POST(BusiAPI+"/SagaBTransIn", common.WrapHandler(sagaBarrierTransIn))
		app.POST(BusiAPI+"/SagaBTransInCompensate", common.WrapHandler(sagaBarrierTransInCompensate))
		app.POST(BusiAPI+"/SagaBTransOut", common.WrapHandler(sagaBarrierTransOut))
		app.POST(BusiAPI+"/SagaBTransOutCompensate", common.WrapHandler(sagaBarrierTransOutCompensate))
	}
	addSample("saga_barrier", func() string {
		dtmcli.Logf("a busi transaction begin")
		req := &TransReq{Amount: 30}
		saga := dtmcli.NewSaga(DtmServer, dtmcli.MustGenGid(DtmServer)).
			Add(Busi+"/SagaBTransOut", Busi+"/SagaBTransOutCompensate", req).
			Add(Busi+"/SagaBTransIn", Busi+"/SagaBTransInCompensate", req)
		dtmcli.Logf("busi trans submit")
		err := saga.Submit()
		dtmcli.FatalIfError(err)
		return saga.Gid
	})
}

func sagaBarrierAdjustBalance(sdb *sql.Tx, uid int, amount int) error {
	_, err := dtmcli.StxExec(sdb, "update dtm_busi.user_account set balance = balance + ? where user_id = ?", amount, uid)
	return err

}

func sagaBarrierTransIn(c *gin.Context) (interface{}, error) {
	req := reqFrom(c)
	if req.TransInResult != "" {
		return req.TransInResult, nil
	}
	barrier := MustBarrierFromGin(c)
	return dtmcli.ResultSuccess, barrier.Call(sdbGet(), func(sdb *sql.Tx) error {
		return sagaBarrierAdjustBalance(sdb, 1, req.Amount)
	})
}

func sagaBarrierTransInCompensate(c *gin.Context) (interface{}, error) {
	barrier := MustBarrierFromGin(c)
	return dtmcli.ResultSuccess, barrier.Call(sdbGet(), func(sdb *sql.Tx) error {
		return sagaBarrierAdjustBalance(sdb, 1, -reqFrom(c).Amount)
	})
}

func sagaBarrierTransOut(c *gin.Context) (interface{}, error) {
	req := reqFrom(c)
	if req.TransInResult != "" {
		return req.TransInResult, nil
	}
	barrier := MustBarrierFromGin(c)
	return dtmcli.ResultSuccess, barrier.Call(sdbGet(), func(sdb *sql.Tx) error {
		return sagaBarrierAdjustBalance(sdb, 2, -req.Amount)
	})
}

func sagaBarrierTransOutCompensate(c *gin.Context) (interface{}, error) {
	barrier := MustBarrierFromGin(c)
	return dtmcli.ResultSuccess, barrier.Call(sdbGet(), func(sdb *sql.Tx) error {
		return sagaBarrierAdjustBalance(sdb, 2, reqFrom(c).Amount)
	})
}
