package logic

import (
	in "wallet_service/internal/app/interfaces"
)

func calcOrderSum(orderData *in.OrderDTO) float32 {
	var sum float32
	for _, v := range orderData.OrderItems {
		sum += v.ProductPrice * float32(v.Count)
	}

	return sum
}
