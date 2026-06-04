package sim

const (
	CommissionRate = 0.0003
	MinCommission  = 5.0
	StampTaxRate   = 0.001
	TransferRate   = 0.00002 // 过户费约万0.2，双向
)

func CalcCommission(amount float64) float64 {
	c := amount * CommissionRate
	if c < MinCommission {
		return round2(MinCommission)
	}
	return round2(c)
}

func CalcStampTax(amount float64) float64 {
	return round2(amount * StampTaxRate)
}

func CalcTransferFee(amount float64) float64 {
	return round2(amount * TransferRate)
}

func round2(v float64) float64 {
	return float64(int64(v*100+0.5)) / 100
}
