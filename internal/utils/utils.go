package utils

import "log"

func HandleBuyTransaction(buyQuant, oldQuant int, oldAvgPrice, currPrice float64) (newQuant int, totalInvested, newAvg, pnl, pnlPercentage, totalAmount float64) {
	newQuant = oldQuant + buyQuant
	totalInvested = float64(oldQuant)*oldAvgPrice + float64(buyQuant)*currPrice
	newAvg = totalInvested / float64(newQuant)
	pnl = float64(newQuant)*currPrice - totalInvested
	pnlPercentage = (pnl / totalInvested) * 100
	totalAmount = float64(buyQuant) * currPrice
	return
}

func HandleSellTransaction(sellQuant, oldQuant int, oldAvgPrice, currPrice float64) (newQuant int, totalInvested, newAvg, pnl, pnlPercentage, totalAmount float64) {
	if sellQuant > oldQuant {
		log.Panic("Trying to sell a stock you don't own niga")
	}
	if sellQuant == oldQuant {
		newQuant = 0
		totalInvested = 0
		newAvg = 0
		pnl = 0
		pnlPercentage = 0
		totalAmount = 0
		return
	}
	newQuant = oldQuant - sellQuant
	totalInvested = float64(newQuant) * oldAvgPrice
	newAvg = oldAvgPrice
	pnl = float64(newQuant)*currPrice - totalInvested
	pnlPercentage = (pnl / totalInvested) * 100
	totalAmount = float64(sellQuant) * currPrice
	return
}
