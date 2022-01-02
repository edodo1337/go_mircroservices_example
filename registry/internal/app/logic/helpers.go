package logic

import (
	in "registry_service/internal/app/interfaces"
)

// Helper func for products prices enrichment.
func enrichOrderItemsDataWithPrices(
	productPricesMap in.ProductPricesMap,
	orderItemsData []*in.MakeOrderItemDTO,
) []*in.NewOrderItemDTO {
	orderItemsDTOs := make([]*in.NewOrderItemDTO, 0, 10)

	for _, item := range orderItemsData {
		val := productPricesMap[item.ProductID]

		orderItemsDTOs = append(orderItemsDTOs, &in.NewOrderItemDTO{
			ProductID:    item.ProductID,
			Count:        item.Count,
			ProductPrice: val,
		})
	}

	return orderItemsDTOs
}
