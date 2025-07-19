package proto

import (
	duelDomain "github.com/peterparker2005/giftduels/apps/service-duel/internal/domain/duel"
	duelv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/duel/v1"
)

func MapFilter(filter *duelDomain.Filter) *duelv1.GetDuelListFilter {
	switch filter.FilterType {
	case duelDomain.FilterTypeAll:
		return &duelv1.GetDuelListFilter{
			FilterType: duelv1.GetDuelListFilter_FILTER_TYPE_ALL,
		}
	case duelDomain.FilterType1v1:
		return &duelv1.GetDuelListFilter{
			FilterType: duelv1.GetDuelListFilter_FILTER_TYPE_1V1,
		}
	case duelDomain.FilterTypeDailyTop:
		return &duelv1.GetDuelListFilter{
			FilterType: duelv1.GetDuelListFilter_FILTER_TYPE_DAILY_TOP,
		}
	case duelDomain.FilterTypeMyDuels:
		return &duelv1.GetDuelListFilter{
			FilterType: duelv1.GetDuelListFilter_FILTER_TYPE_MY_DUELS,
		}
	default:
		return &duelv1.GetDuelListFilter{
			FilterType: duelv1.GetDuelListFilter_FILTER_TYPE_UNSPECIFIED,
		}
	}
}

func MapFilterType(filterType duelv1.GetDuelListFilter_FilterType) (duelDomain.FilterType, error) {
	switch filterType {
	case duelv1.GetDuelListFilter_FILTER_TYPE_ALL:
		return duelDomain.FilterTypeAll, nil
	case duelv1.GetDuelListFilter_FILTER_TYPE_1V1:
		return duelDomain.FilterType1v1, nil
	case duelv1.GetDuelListFilter_FILTER_TYPE_DAILY_TOP:
		return duelDomain.FilterTypeDailyTop, nil
	case duelv1.GetDuelListFilter_FILTER_TYPE_MY_DUELS:
		return duelDomain.FilterTypeMyDuels, nil
	case duelv1.GetDuelListFilter_FILTER_TYPE_UNSPECIFIED:
		return "", ErrInvalidFilterType
	default:
		return "", ErrInvalidFilterType
	}
}
