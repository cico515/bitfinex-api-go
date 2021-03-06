package bitfinex

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/bitfinexcom/bitfinex-api-go/pkg/convert"
)

// Prefixes for available pairs
const (
	FundingPrefix = "f"
	TradingPrefix = "t"
)

var (
	ErrNotFound = errors.New("not found")
)

// Candle resolutions
const (
	OneMinute      CandleResolution = "1m"
	FiveMinutes    CandleResolution = "5m"
	FifteenMinutes CandleResolution = "15m"
	ThirtyMinutes  CandleResolution = "30m"
	OneHour        CandleResolution = "1h"
	ThreeHours     CandleResolution = "3h"
	SixHours       CandleResolution = "6h"
	TwelveHours    CandleResolution = "12h"
	OneDay         CandleResolution = "1D"
	OneWeek        CandleResolution = "7D"
	TwoWeeks       CandleResolution = "14D"
	OneMonth       CandleResolution = "1M"
)

type Mts int64
type SortOrder int

const (
	OldestFirst SortOrder = 1
	NewestFirst SortOrder = -1
)

type QueryLimit int

const QueryLimitMax QueryLimit = 1000

func CandleResolutionFromString(str string) (CandleResolution, error) {
	switch str {
	case string(OneMinute):
		return OneMinute, nil
	case string(FiveMinutes):
		return FiveMinutes, nil
	case string(FifteenMinutes):
		return FifteenMinutes, nil
	case string(ThirtyMinutes):
		return ThirtyMinutes, nil
	case string(OneHour):
		return OneHour, nil
	case string(ThreeHours):
		return ThreeHours, nil
	case string(SixHours):
		return SixHours, nil
	case string(TwelveHours):
		return TwelveHours, nil
	case string(OneDay):
		return OneDay, nil
	case string(OneWeek):
		return OneWeek, nil
	case string(TwoWeeks):
		return TwoWeeks, nil
	case string(OneMonth):
		return OneMonth, nil
	}
	return OneMinute, fmt.Errorf("could not convert string to resolution: %s", str)
}

type PermissionType string

const (
	PermissionRead  = "r"
	PermissionWrite = "w"
)

// private type--cannot instantiate.
type candleResolution string

// CandleResolution provides a typed set of resolutions for candle subscriptions.
type CandleResolution candleResolution

// Order sides
const (
	Bid   OrderSide = 1
	Ask   OrderSide = 2
	Long  OrderSide = 1
	Short OrderSide = 2
)

// Settings flags

const (
	Dec_s     int = 9
	Time_s    int = 32
	Timestamp int = 32768
	Seq_all   int = 65536
	Checksum  int = 131072
)

type orderSide byte

// OrderSide provides a typed set of order sides.
type OrderSide orderSide

// Book precision levels
const (
	// Aggregate precision levels
	Precision0 BookPrecision = "P0"
	Precision2 BookPrecision = "P2"
	Precision1 BookPrecision = "P1"
	Precision3 BookPrecision = "P3"
	// Raw precision
	PrecisionRawBook BookPrecision = "R0"
)

// private type
type bookPrecision string

// BookPrecision provides a typed book precision level.
type BookPrecision bookPrecision

const (
	// FrequencyRealtime book frequency gives updates as they occur in real-time.
	FrequencyRealtime BookFrequency = "F0"
	// FrequencyTwoPerSecond delivers two book updates per second.
	FrequencyTwoPerSecond BookFrequency = "F1"
	// PriceLevelDefault provides a constant default price level for book subscriptions.
	PriceLevelDefault int = 25
)

type bookFrequency string

// BookFrequency provides a typed book frequency.
type BookFrequency bookFrequency

const (
	OrderFlagHidden   int = 64
	OrderFlagClose    int = 512
	OrderFlagPostOnly int = 4096
	OrderFlagOCO      int = 16384
)

// OrderNewRequest represents an order to be posted to the bitfinex websocket
// service.
type OrderNewRequest struct {
	GID           int64                  `json:"gid"`
	CID           int64                  `json:"cid"`
	Type          string                 `json:"type"`
	Symbol        string                 `json:"symbol"`
	Amount        float64                `json:"amount,string"`
	Price         float64                `json:"price,string"`
	Leverage      int64                  `json:"lev,omitempty"`
	PriceTrailing float64                `json:"price_trailing,string,omitempty"`
	PriceAuxLimit float64                `json:"price_aux_limit,string,omitempty"`
	PriceOcoStop  float64                `json:"price_oco_stop,string,omitempty"`
	Hidden        bool                   `json:"hidden,omitempty"`
	PostOnly      bool                   `json:"postonly,omitempty"`
	Close         bool                   `json:"close,omitempty"`
	OcoOrder      bool                   `json:"oco_order,omitempty"`
	TimeInForce   string                 `json:"tif,omitempty"`
	AffiliateCode string                 `json:"-"`
	Meta          map[string]interface{} `json:"meta,omitempty"`
}

type OrderMeta struct {
	AffiliateCode string `json:"aff_code,string,omitempty"`
}

// MarshalJSON converts the order object into the format required by the bitfinex
// websocket service.
func (o *OrderNewRequest) MarshalJSON() ([]byte, error) {
	jsonOrder, err := o.ToJSON()
	if err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf("[0, \"on\", null, %s]", string(jsonOrder))), nil
}

// EnrichedPayload returns enriched representation of order struct for submission
func (o *OrderNewRequest) EnrichedPayload() interface{} {
	pld := struct {
		GID           int64                  `json:"gid"`
		CID           int64                  `json:"cid"`
		Type          string                 `json:"type"`
		Symbol        string                 `json:"symbol"`
		Amount        float64                `json:"amount,string"`
		Price         float64                `json:"price,string"`
		Leverage      int64                  `json:"lev,omitempty"`
		PriceTrailing float64                `json:"price_trailing,string,omitempty"`
		PriceAuxLimit float64                `json:"price_aux_limit,string,omitempty"`
		PriceOcoStop  float64                `json:"price_oco_stop,string,omitempty"`
		TimeInForce   string                 `json:"tif,omitempty"`
		Flags         int                    `json:"flags,omitempty"`
		Meta          map[string]interface{} `json:"meta,omitempty"`
	}{
		GID:           o.GID,
		CID:           o.CID,
		Type:          o.Type,
		Symbol:        o.Symbol,
		Amount:        o.Amount,
		Price:         o.Price,
		Leverage:      o.Leverage,
		PriceTrailing: o.PriceTrailing,
		PriceAuxLimit: o.PriceAuxLimit,
		PriceOcoStop:  o.PriceOcoStop,
		TimeInForce:   o.TimeInForce,
	}

	if o.Hidden {
		pld.Flags = pld.Flags + OrderFlagHidden
	}

	if o.PostOnly {
		pld.Flags = pld.Flags + OrderFlagPostOnly
	}

	if o.OcoOrder {
		pld.Flags = pld.Flags + OrderFlagOCO
	}

	if o.Close {
		pld.Flags = pld.Flags + OrderFlagClose
	}

	if o.Meta == nil {
		pld.Meta = make(map[string]interface{})
	}

	if o.AffiliateCode != "" {
		pld.Meta["aff_code"] = o.AffiliateCode
	}

	return pld
}

func (o *OrderNewRequest) ToJSON() ([]byte, error) {
	return json.Marshal(o.EnrichedPayload())
}

type OrderUpdateRequest struct {
	ID            int64                  `json:"id"`
	GID           int64                  `json:"gid,omitempty"`
	Price         float64                `json:"price,string,omitempty"`
	Amount        float64                `json:"amount,string,omitempty"`
	Leverage      int64                  `json:"lev,omitempty"`
	Delta         float64                `json:"delta,string,omitempty"`
	PriceTrailing float64                `json:"price_trailing,string,omitempty"`
	PriceAuxLimit float64                `json:"price_aux_limit,string,omitempty"`
	Hidden        bool                   `json:"hidden,omitempty"`
	PostOnly      bool                   `json:"postonly,omitempty"`
	TimeInForce   string                 `json:"tif,omitempty"`
	Meta          map[string]interface{} `json:"meta,omitempty"`
}

// MarshalJSON converts the order object into the format required by the bitfinex
// websocket service.
func (o *OrderUpdateRequest) MarshalJSON() ([]byte, error) {
	aux, err := o.ToJSON()
	if err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf("[0, \"ou\", null, %s]", string(aux))), nil
}

func (o *OrderUpdateRequest) EnrichedPayload() interface{} {
	pld := struct {
		ID            int64                  `json:"id"`
		GID           int64                  `json:"gid,omitempty"`
		Price         float64                `json:"price,string,omitempty"`
		Amount        float64                `json:"amount,string,omitempty"`
		Leverage      int64                  `json:"lev,omitempty"`
		Delta         float64                `json:"delta,string,omitempty"`
		PriceTrailing float64                `json:"price_trailing,string,omitempty"`
		PriceAuxLimit float64                `json:"price_aux_limit,string,omitempty"`
		Hidden        bool                   `json:"hidden,omitempty"`
		PostOnly      bool                   `json:"postonly,omitempty"`
		TimeInForce   string                 `json:"tif,omitempty"`
		Flags         int                    `json:"flags,omitempty"`
		Meta          map[string]interface{} `json:"meta,omitempty"`
	}{
		ID:            o.ID,
		GID:           o.GID,
		Amount:        o.Amount,
		Leverage:      o.Leverage,
		Price:         o.Price,
		PriceTrailing: o.PriceTrailing,
		PriceAuxLimit: o.PriceAuxLimit,
		Delta:         o.Delta,
		TimeInForce:   o.TimeInForce,
	}

	if o.Meta == nil {
		pld.Meta = make(map[string]interface{})
	}

	if o.Hidden {
		pld.Flags = pld.Flags + OrderFlagHidden
	}

	if o.PostOnly {
		pld.Flags = pld.Flags + OrderFlagPostOnly
	}

	return pld
}

func (o *OrderUpdateRequest) ToJSON() ([]byte, error) {
	return json.Marshal(o.EnrichedPayload())
}

// OrderCancelRequest represents an order cancel request.
// An order can be cancelled using the internal ID or a
// combination of Client ID (CID) and the daten for the given
// CID.
type OrderCancelRequest struct {
	ID      int64  `json:"id,omitempty"`
	CID     int64  `json:"cid,omitempty"`
	CIDDate string `json:"cid_date,omitempty"`
}

func (o *OrderCancelRequest) ToJSON() ([]byte, error) {
	aux := struct {
		ID      int64  `json:"id,omitempty"`
		CID     int64  `json:"cid,omitempty"`
		CIDDate string `json:"cid_date,omitempty"`
	}{
		ID:      o.ID,
		CID:     o.CID,
		CIDDate: o.CIDDate,
	}

	return json.Marshal(aux)
}

// MarshalJSON converts the order cancel object into the format required by the
// bitfinex websocket service.
func (o *OrderCancelRequest) MarshalJSON() ([]byte, error) {
	aux, err := o.ToJSON()
	if err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf("[0, \"oc\", null, %s]", string(aux))), nil
}

// TODO: MultiOrderCancelRequest represents an order cancel request.

type Heartbeat struct {
	//ChannelIDs []int64
}

// OrderType represents the types orders the bitfinex platform can handle.
type OrderType string

const (
	OrderTypeMarket               = "MARKET"
	OrderTypeExchangeMarket       = "EXCHANGE MARKET"
	OrderTypeLimit                = "LIMIT"
	OrderTypeExchangeLimit        = "EXCHANGE LIMIT"
	OrderTypeStop                 = "STOP"
	OrderTypeExchangeStop         = "EXCHANGE STOP"
	OrderTypeTrailingStop         = "TRAILING STOP"
	OrderTypeExchangeTrailingStop = "EXCHANGE TRAILING STOP"
	OrderTypeFOK                  = "FOK"
	OrderTypeExchangeFOK          = "EXCHANGE FOK"
	OrderTypeStopLimit            = "STOP LIMIT"
	OrderTypeExchangeStopLimit    = "EXCHANGE STOP LIMIT"
)

// OrderStatus represents the possible statuses an order can be in.
type OrderStatus string

const (
	OrderStatusActive          OrderStatus = "ACTIVE"
	OrderStatusExecuted        OrderStatus = "EXECUTED"
	OrderStatusPartiallyFilled OrderStatus = "PARTIALLY FILLED"
	OrderStatusCanceled        OrderStatus = "CANCELED"
)

// Order as returned from the bitfinex websocket service.
type Order struct {
	ID            int64
	GID           int64
	CID           int64
	Symbol        string
	MTSCreated    int64
	MTSUpdated    int64
	Amount        float64
	AmountOrig    float64
	Type          string
	TypePrev      string
	MTSTif        int64
	Flags         int64
	Status        OrderStatus
	Price         float64
	PriceAvg      float64
	PriceTrailing float64
	PriceAuxLimit float64
	Notify        bool
	Hidden        bool
	PlacedID      int64
	Meta          map[string]interface{}
}

// NewOrderFromRaw takes the raw list of values as returned from the websocket
// service and tries to convert it into an Order.
func NewOrderFromRaw(raw []interface{}) (o *Order, err error) {
	if len(raw) == 12 {
		o = &Order{
			ID:         int64(convert.F64ValOrZero(raw[0])),
			Symbol:     convert.SValOrEmpty(raw[1]),
			Amount:     convert.F64ValOrZero(raw[2]),
			AmountOrig: convert.F64ValOrZero(raw[3]),
			Type:       convert.SValOrEmpty(raw[4]),
			Status:     OrderStatus(convert.SValOrEmpty(raw[5])),
			Price:      convert.F64ValOrZero(raw[6]),
			PriceAvg:   convert.F64ValOrZero(raw[7]),
			MTSUpdated: convert.I64ValOrZero(raw[8]),
			// 3 trailing zeroes, what do they map to?
		}
	} else if len(raw) < 26 {
		return o, fmt.Errorf("data slice too short for order: %#v", raw)
	} else {
		o = &Order{
			ID:            int64(convert.F64ValOrZero(raw[0])),
			GID:           int64(convert.F64ValOrZero(raw[1])),
			CID:           int64(convert.F64ValOrZero(raw[2])),
			Symbol:        convert.SValOrEmpty(raw[3]),
			MTSCreated:    int64(convert.F64ValOrZero(raw[4])),
			MTSUpdated:    int64(convert.F64ValOrZero(raw[5])),
			Amount:        convert.F64ValOrZero(raw[6]),
			AmountOrig:    convert.F64ValOrZero(raw[7]),
			Type:          convert.SValOrEmpty(raw[8]),
			TypePrev:      convert.SValOrEmpty(raw[9]),
			MTSTif:        int64(convert.F64ValOrZero(raw[10])),
			Flags:         convert.I64ValOrZero(raw[12]),
			Status:        OrderStatus(convert.SValOrEmpty(raw[13])),
			Price:         convert.F64ValOrZero(raw[16]),
			PriceAvg:      convert.F64ValOrZero(raw[17]),
			PriceTrailing: convert.F64ValOrZero(raw[18]),
			PriceAuxLimit: convert.F64ValOrZero(raw[19]),
			Notify:        convert.BValOrFalse(raw[23]),
			Hidden:        convert.BValOrFalse(raw[24]),
			PlacedID:      convert.I64ValOrZero(raw[25]),
		}
	}
	if len(raw) >= 31 {
		o.Meta = convert.SiMapOrEmpty(raw[31])
	}
	return o, nil
}

// OrderSnapshotFromRaw takes a raw list of values as returned from the websocket
// service and tries to convert it into an OrderSnapshot.
func NewOrderSnapshotFromRaw(raw []interface{}) (s *OrderSnapshot, err error) {
	if len(raw) == 0 {
		return
	}

	os := make([]*Order, 0)
	switch raw[0].(type) {
	case []interface{}:
		for _, v := range raw {
			if l, ok := v.([]interface{}); ok {
				o, err := NewOrderFromRaw(l)
				if err != nil {
					return s, err
				}
				os = append(os, o)
			}
		}
	default:
		return s, fmt.Errorf("not an order snapshot")
	}
	s = &OrderSnapshot{Snapshot: os}

	return
}

// OrderSnapshot is a collection of Orders that would usually be sent on
// inital connection.
type OrderSnapshot struct {
	Snapshot []*Order
}

// OrderUpdate is an Order that gets sent out after every change to an
// order.
type OrderUpdate Order

// OrderNew gets sent out after an Order was created successfully.
type OrderNew Order

// OrderCancel gets sent out after an Order was cancelled successfully.
type OrderCancel Order

type PositionStatus string

const (
	PositionStatusActive PositionStatus = "ACTIVE"
	PositionStatusClosed PositionStatus = "CLOSED"
)

type Position struct {
	Id                   int64
	Symbol               string
	Status               PositionStatus
	Amount               float64
	BasePrice            float64
	MarginFunding        float64
	MarginFundingType    int64
	ProfitLoss           float64
	ProfitLossPercentage float64
	LiquidationPrice     float64
	Leverage             float64
}

func NewPositionFromRaw(raw []interface{}) (o *Position, err error) {
	if len(raw) == 6 {
		o = &Position{
			Symbol:            convert.SValOrEmpty(raw[0]),
			Status:            PositionStatus(convert.SValOrEmpty(raw[1])),
			Amount:            convert.F64ValOrZero(raw[2]),
			BasePrice:         convert.F64ValOrZero(raw[3]),
			MarginFunding:     convert.F64ValOrZero(raw[4]),
			MarginFundingType: convert.I64ValOrZero(raw[5]),
		}
	} else if len(raw) < 10 {
		return o, fmt.Errorf("data slice too short for position: %#v", raw)
	} else if len(raw) == 10 {
		o = &Position{
			Symbol:               convert.SValOrEmpty(raw[0]),
			Status:               PositionStatus(convert.SValOrEmpty(raw[1])),
			Amount:               convert.F64ValOrZero(raw[2]),
			BasePrice:            convert.F64ValOrZero(raw[3]),
			MarginFunding:        convert.F64ValOrZero(raw[4]),
			MarginFundingType:    convert.I64ValOrZero(raw[5]),
			ProfitLoss:           convert.F64ValOrZero(raw[6]),
			ProfitLossPercentage: convert.F64ValOrZero(raw[7]),
			LiquidationPrice:     convert.F64ValOrZero(raw[8]),
			Leverage:             convert.F64ValOrZero(raw[9]),
		}
	} else {
		o = &Position{
			Symbol:               convert.SValOrEmpty(raw[0]),
			Status:               PositionStatus(convert.SValOrEmpty(raw[1])),
			Amount:               convert.F64ValOrZero(raw[2]),
			BasePrice:            convert.F64ValOrZero(raw[3]),
			MarginFunding:        convert.F64ValOrZero(raw[4]),
			MarginFundingType:    convert.I64ValOrZero(raw[5]),
			ProfitLoss:           convert.F64ValOrZero(raw[6]),
			ProfitLossPercentage: convert.F64ValOrZero(raw[7]),
			LiquidationPrice:     convert.F64ValOrZero(raw[8]),
			Leverage:             convert.F64ValOrZero(raw[9]),
			Id:                   int64(convert.F64ValOrZero(raw[11])),
		}
	}
	return
}

type PositionSnapshot struct {
	Snapshot []*Position
}
type PositionNew Position
type PositionUpdate Position
type PositionCancel Position

func NewPositionSnapshotFromRaw(raw []interface{}) (s *PositionSnapshot, err error) {
	if len(raw) == 0 {
		return
	}

	ps := make([]*Position, 0)
	switch raw[0].(type) {
	case []interface{}:
		for _, v := range raw {
			if l, ok := v.([]interface{}); ok {
				p, err := NewPositionFromRaw(l)
				if err != nil {
					return s, err
				}
				ps = append(ps, p)
			}
		}
	default:
		return s, fmt.Errorf("not a position snapshot")
	}
	s = &PositionSnapshot{Snapshot: ps}

	return
}

type ClaimPositionRequest struct {
	Id int64
}

func (o *ClaimPositionRequest) ToJSON() ([]byte, error) {
	aux := struct {
		Id int64 `json:"id"`
	}{
		Id: o.Id,
	}
	return json.Marshal(aux)
}

// Trade represents a trade on the public data feed.
type Trade struct {
	Pair   string
	ID     int64
	MTS    int64
	Amount float64
	Price  float64
	Side   OrderSide
}

func NewTradeFromRaw(pair string, raw []interface{}) (o *Trade, err error) {
	if len(raw) < 4 {
		return o, fmt.Errorf("data slice too short for trade: %#v", raw)
	}

	amt := convert.F64ValOrZero(raw[2])
	var side OrderSide
	if amt > 0 {
		side = Bid
	} else {
		side = Ask
	}

	o = &Trade{
		Pair:   pair,
		ID:     convert.I64ValOrZero(raw[0]),
		MTS:    convert.I64ValOrZero(raw[1]),
		Amount: math.Abs(amt),
		Price:  convert.F64ValOrZero(raw[3]),
		Side:   side,
	}

	return
}

type TradeSnapshot struct {
	Snapshot []*Trade
}

func NewTradeSnapshotFromRaw(pair string, raw [][]float64) (*TradeSnapshot, error) {
	if len(raw) <= 0 {
		return nil, fmt.Errorf("data slice is too short for trade snapshot: %#v", raw)
	}
	snapshot := make([]*Trade, 0)
	for _, flt := range raw {
		t, err := NewTradeFromRaw(pair, ToInterface(flt))
		if err == nil {
			snapshot = append(snapshot, t)
		}
	}

	return &TradeSnapshot{Snapshot: snapshot}, nil
}

// TradeExecutionUpdate represents a full update to a trade on the private data feed.  Following a TradeExecution,
// TradeExecutionUpdates include additional details, e.g. the trade's execution ID (TradeID).
type TradeExecutionUpdate struct {
	ID          int64
	Pair        string
	MTS         int64
	OrderID     int64
	ExecAmount  float64
	ExecPrice   float64
	OrderType   string
	OrderPrice  float64
	Maker       int
	Fee         float64
	FeeCurrency string
}

// public trade update just looks like a trade
func NewTradeExecutionUpdateFromRaw(raw []interface{}) (o *TradeExecutionUpdate, err error) {
	if len(raw) == 4 {
		o = &TradeExecutionUpdate{
			ID:         convert.I64ValOrZero(raw[0]),
			MTS:        convert.I64ValOrZero(raw[1]),
			ExecAmount: convert.F64ValOrZero(raw[2]),
			ExecPrice:  convert.F64ValOrZero(raw[3]),
		}
		return
	}
	if len(raw) == 11 {
		o = &TradeExecutionUpdate{
			ID:          convert.I64ValOrZero(raw[0]),
			Pair:        convert.SValOrEmpty(raw[1]),
			MTS:         convert.I64ValOrZero(raw[2]),
			OrderID:     convert.I64ValOrZero(raw[3]),
			ExecAmount:  convert.F64ValOrZero(raw[4]),
			ExecPrice:   convert.F64ValOrZero(raw[5]),
			OrderType:   convert.SValOrEmpty(raw[6]),
			OrderPrice:  convert.F64ValOrZero(raw[7]),
			Maker:       convert.IValOrZero(raw[8]),
			Fee:         convert.F64ValOrZero(raw[9]),
			FeeCurrency: convert.SValOrEmpty(raw[10]),
		}
		return
	}
	return o, fmt.Errorf("data slice too short for trade update: %#v", raw)
}

type TradeExecutionUpdateSnapshot struct {
	Snapshot []*TradeExecutionUpdate
}
type HistoricalTradeSnapshot TradeExecutionUpdateSnapshot

func NewTradeExecutionUpdateSnapshotFromRaw(raw []interface{}) (s *TradeExecutionUpdateSnapshot, err error) {
	if len(raw) == 0 {
		return
	}
	ts := make([]*TradeExecutionUpdate, 0)
	switch raw[0].(type) {
	case []interface{}:
		for _, v := range raw {
			if l, ok := v.([]interface{}); ok {
				t, err := NewTradeExecutionUpdateFromRaw(l)
				if err != nil {
					return s, err
				}
				ts = append(ts, t)
			}
		}
	default:
		return s, fmt.Errorf("not a trade snapshot: %#v", raw)
	}
	s = &TradeExecutionUpdateSnapshot{Snapshot: ts}

	return
}

// TradeExecution represents the first message receievd for a trade on the private data feed.
type TradeExecution struct {
	ID         int64
	Pair       string
	MTS        int64
	OrderID    int64
	Amount     float64
	Price      float64
	OrderType  string
	OrderPrice float64
	Maker      int
}

func NewTradeExecutionFromRaw(raw []interface{}) (o *TradeExecution, err error) {
	if len(raw) < 6 {
		log.Printf("[ERROR] not enough members (%d, need at least 6) for trade execution: %#v", len(raw), raw)
		return o, fmt.Errorf("data slice too short for trade execution: %#v", raw)
	}

	// trade executions sometimes omit order type, price, and maker flag
	o = &TradeExecution{
		ID:      convert.I64ValOrZero(raw[0]),
		Pair:    convert.SValOrEmpty(raw[1]),
		MTS:     convert.I64ValOrZero(raw[2]),
		OrderID: convert.I64ValOrZero(raw[3]),
		Amount:  convert.F64ValOrZero(raw[4]),
		Price:   convert.F64ValOrZero(raw[5]),
	}

	if len(raw) >= 9 {
		o.OrderType = convert.SValOrEmpty(raw[6])
		o.OrderPrice = convert.F64ValOrZero(raw[7])
		o.Maker = convert.IValOrZero(raw[8])
	}

	return
}

type Wallet struct {
	Type              string
	Currency          string
	Balance           float64
	UnsettledInterest float64
	BalanceAvailable  float64
}

func NewWalletFromRaw(raw []interface{}) (o *Wallet, err error) {
	if len(raw) == 4 {
		o = &Wallet{
			Type:              convert.SValOrEmpty(raw[0]),
			Currency:          convert.SValOrEmpty(raw[1]),
			Balance:           convert.F64ValOrZero(raw[2]),
			UnsettledInterest: convert.F64ValOrZero(raw[3]),
		}
	} else if len(raw) < 5 {
		return o, fmt.Errorf("data slice too short for wallet: %#v", raw)
	} else {
		o = &Wallet{
			Type:              convert.SValOrEmpty(raw[0]),
			Currency:          convert.SValOrEmpty(raw[1]),
			Balance:           convert.F64ValOrZero(raw[2]),
			UnsettledInterest: convert.F64ValOrZero(raw[3]),
			BalanceAvailable:  convert.F64ValOrZero(raw[4]),
		}
	}
	return
}

type WalletUpdate Wallet
type WalletSnapshot struct {
	Snapshot []*Wallet
}

func NewWalletSnapshotFromRaw(raw []interface{}) (s *WalletSnapshot, err error) {
	if len(raw) == 0 {
		return
	}

	ws := make([]*Wallet, 0)
	switch raw[0].(type) {
	case []interface{}:
		for _, v := range raw {
			if l, ok := v.([]interface{}); ok {
				o, err := NewWalletFromRaw(l)
				if err != nil {
					return s, err
				}
				ws = append(ws, o)
			}
		}
	default:
		return s, fmt.Errorf("not an wallet snapshot")
	}
	s = &WalletSnapshot{Snapshot: ws}

	return
}

type BalanceInfo struct {
	TotalAUM float64
	NetAUM   float64
	/*WalletType string
	Currency   string*/
}

func NewBalanceInfoFromRaw(raw []interface{}) (o *BalanceInfo, err error) {
	if len(raw) < 2 {
		return o, fmt.Errorf("data slice too short for balance info: %#v", raw)
	}

	o = &BalanceInfo{
		TotalAUM: convert.F64ValOrZero(raw[0]),
		NetAUM:   convert.F64ValOrZero(raw[1]),
		/*WalletType: convert.SValOrEmpty(raw[2]),
		Currency:   convert.SValOrEmpty(raw[3]),*/
	}

	return
}

type BalanceUpdate BalanceInfo

// marginInfoFromRaw returns either a MarginInfoBase or MarginInfoUpdate, since
// the Margin Info is split up into a base and per symbol parts.
func NewMarginInfoFromRaw(raw []interface{}) (o interface{}, err error) {
	if len(raw) < 2 {
		return o, fmt.Errorf("data slice too short for margin info base: %#v", raw)
	}

	typ, ok := raw[0].(string)
	if !ok {
		return o, fmt.Errorf("expected margin info type in first position for margin info but got %#v", raw)
	}

	if len(raw) == 2 && typ == "base" { // This should be ["base", [...]]
		data, ok := raw[1].([]interface{})
		if !ok {
			return o, fmt.Errorf("expected margin info array in second position for margin info but got %#v", raw)
		}

		return NewMarginInfoBaseFromRaw(data)
	} else if len(raw) == 3 && typ == "sym" { // This should be ["sym", SYMBOL, [...]]
		symbol, ok := raw[1].(string)
		if !ok {
			return o, fmt.Errorf("expected margin info symbol in second position for margin info update but got %#v", raw)
		}

		data, ok := raw[2].([]interface{})
		if !ok {
			return o, fmt.Errorf("expected margin info array in third position for margin info update but got %#v", raw)
		}

		return NewMarginInfoUpdateFromRaw(symbol, data)
	}

	return nil, fmt.Errorf("invalid margin info type in %#v", raw)
}

type MarginInfoUpdate struct {
	Symbol          string
	TradableBalance float64
}

func NewMarginInfoUpdateFromRaw(symbol string, raw []interface{}) (o *MarginInfoUpdate, err error) {
	if len(raw) < 1 {
		return o, fmt.Errorf("data slice too short for margin info update: %#v", raw)
	}

	o = &MarginInfoUpdate{
		Symbol:          symbol,
		TradableBalance: convert.F64ValOrZero(raw[0]),
	}

	return
}

type MarginInfoBase struct {
	UserProfitLoss float64
	UserSwaps      float64
	MarginBalance  float64
	MarginNet      float64
}

func NewMarginInfoBaseFromRaw(raw []interface{}) (o *MarginInfoBase, err error) {
	if len(raw) < 4 {
		return o, fmt.Errorf("data slice too short for margin info base: %#v", raw)
	}

	o = &MarginInfoBase{
		UserProfitLoss: convert.F64ValOrZero(raw[0]),
		UserSwaps:      convert.F64ValOrZero(raw[1]),
		MarginBalance:  convert.F64ValOrZero(raw[2]),
		MarginNet:      convert.F64ValOrZero(raw[3]),
	}

	return
}

type FundingInfo struct {
	Symbol       string
	YieldLoan    float64
	YieldLend    float64
	DurationLoan float64
	DurationLend float64
}

func NewFundingInfoFromRaw(raw []interface{}) (o *FundingInfo, err error) {
	if len(raw) < 3 { // "sym", symbol, data
		return o, fmt.Errorf("data slice too short for funding info: %#v", raw)
	}

	sym, ok := raw[1].(string)
	if !ok {
		return o, fmt.Errorf("expected symbol in second position of funding info: %v", raw)
	}

	data, ok := raw[2].([]interface{})
	if !ok {
		return o, fmt.Errorf("expected list in third position of funding info: %v", raw)
	}

	if len(data) < 4 {
		return o, fmt.Errorf("data too short: %#v", data)
	}

	o = &FundingInfo{
		Symbol:       sym,
		YieldLoan:    convert.F64ValOrZero(data[0]),
		YieldLend:    convert.F64ValOrZero(data[1]),
		DurationLoan: convert.F64ValOrZero(data[2]),
		DurationLend: convert.F64ValOrZero(data[3]),
	}

	return
}

type OfferStatus string

const (
	OfferStatusActive          OfferStatus = "ACTIVE"
	OfferStatusExecuted        OfferStatus = "EXECUTED"
	OfferStatusPartiallyFilled OfferStatus = "PARTIALLY FILLED"
	OfferStatusCanceled        OfferStatus = "CANCELED"
)

type FundingOfferCancelRequest struct {
	Id int64
}

func (o *FundingOfferCancelRequest) ToJSON() ([]byte, error) {
	aux := struct {
		Id int64 `json:"id"`
	}{
		Id: o.Id,
	}
	return json.Marshal(aux)
}

// MarshalJSON converts the order cancel object into the format required by the
// bitfinex websocket service.
func (o *FundingOfferCancelRequest) MarshalJSON() ([]byte, error) {
	aux, err := o.ToJSON()
	if err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf("[0, \"foc\", null, %s]", string(aux))), nil
}

type FundingOfferRequest struct {
	Type   string
	Symbol string
	Amount float64
	Rate   float64
	Period int64
	Hidden bool
}

func (o *FundingOfferRequest) ToJSON() ([]byte, error) {
	aux := struct {
		Type   string  `json:"type"`
		Symbol string  `json:"symbol"`
		Amount float64 `json:"amount,string"`
		Rate   float64 `json:"rate,string"`
		Period int64   `json:"period"`
		Flags  int     `json:"flags,omitempty"`
	}{
		Type:   o.Type,
		Symbol: o.Symbol,
		Amount: o.Amount,
		Rate:   o.Rate,
		Period: o.Period,
	}
	if o.Hidden {
		aux.Flags = aux.Flags + OrderFlagHidden
	}
	return json.Marshal(aux)
}

// MarshalJSON converts the order cancel object into the format required by the
// bitfinex websocket service.
func (o *FundingOfferRequest) MarshalJSON() ([]byte, error) {
	aux, err := o.ToJSON()
	if err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf("[0, \"fon\", null, %s]", string(aux))), nil
}

type Offer struct {
	ID         int64
	Symbol     string
	MTSCreated int64
	MTSUpdated int64
	Amount     float64
	AmountOrig float64
	Type       string
	Flags      interface{}
	Status     OfferStatus
	Rate       float64
	Period     int64
	Notify     bool
	Hidden     bool
	Insure     bool
	Renew      bool
	RateReal   float64
}

func NewOfferFromRaw(raw []interface{}) (o *Offer, err error) {
	if len(raw) < 21 {
		return o, fmt.Errorf("data slice too short for offer: %#v", raw)
	}

	o = &Offer{
		ID:         convert.I64ValOrZero(raw[0]),
		Symbol:     convert.SValOrEmpty(raw[1]),
		MTSCreated: convert.I64ValOrZero(raw[2]),
		MTSUpdated: convert.I64ValOrZero(raw[3]),
		Amount:     convert.F64ValOrZero(raw[4]),
		AmountOrig: convert.F64ValOrZero(raw[5]),
		Type:       convert.SValOrEmpty(raw[6]),
		Flags:      raw[9],
		Status:     OfferStatus(convert.SValOrEmpty(raw[10])),
		Rate:       convert.F64ValOrZero(raw[14]),
		Period:     convert.I64ValOrZero(raw[15]),
		Notify:     convert.BValOrFalse(raw[16]),
		Hidden:     convert.BValOrFalse(raw[17]),
		Insure:     convert.BValOrFalse(raw[18]),
		Renew:      convert.BValOrFalse(raw[19]),
		RateReal:   convert.F64ValOrZero(raw[20]),
	}

	return
}

type FundingOfferNew Offer
type FundingOfferUpdate Offer
type FundingOfferCancel Offer
type FundingOfferSnapshot struct {
	Snapshot []*Offer
}

func NewFundingOfferSnapshotFromRaw(raw []interface{}) (snap *FundingOfferSnapshot, err error) {
	if len(raw) == 0 {
		return
	}

	fos := make([]*Offer, 0)
	switch raw[0].(type) {
	case []interface{}:
		for _, v := range raw {
			if l, ok := v.([]interface{}); ok {
				o, err := NewOfferFromRaw(l)
				if err != nil {
					return snap, err
				}
				fos = append(fos, o)
			}
		}
	default:
		return snap, fmt.Errorf("not a funding offer snapshot")
	}

	snap = &FundingOfferSnapshot{
		Snapshot: fos,
	}

	return
}

type HistoricalOffer Offer

type CreditStatus string

const (
	CreditStatusActive          CreditStatus = "ACTIVE"
	CreditStatusExecuted        CreditStatus = "EXECUTED"
	CreditStatusPartiallyFilled CreditStatus = "PARTIALLY FILLED"
	CreditStatusCanceled        CreditStatus = "CANCELED"
)

type Credit struct {
	ID            int64
	Symbol        string
	Side          string
	MTSCreated    int64
	MTSUpdated    int64
	Amount        float64
	Flags         interface{}
	Status        CreditStatus
	Rate          float64
	Period        int64
	MTSOpened     int64
	MTSLastPayout int64
	Notify        bool
	Hidden        bool
	Insure        bool
	Renew         bool
	RateReal      float64
	NoClose       bool
	PositionPair  string
}

func NewCreditFromRaw(raw []interface{}) (o *Credit, err error) {
	if len(raw) < 22 {
		return o, fmt.Errorf("data slice too short for offer: %#v", raw)
	}

	o = &Credit{
		ID:            convert.I64ValOrZero(raw[0]),
		Symbol:        convert.SValOrEmpty(raw[1]),
		Side:          convert.SValOrEmpty(raw[2]),
		MTSCreated:    convert.I64ValOrZero(raw[3]),
		MTSUpdated:    convert.I64ValOrZero(raw[4]),
		Amount:        convert.F64ValOrZero(raw[5]),
		Flags:         raw[6],
		Status:        CreditStatus(convert.SValOrEmpty(raw[7])),
		Rate:          convert.F64ValOrZero(raw[11]),
		Period:        convert.I64ValOrZero(raw[12]),
		MTSOpened:     convert.I64ValOrZero(raw[13]),
		MTSLastPayout: convert.I64ValOrZero(raw[14]),
		Notify:        convert.BValOrFalse(raw[15]),
		Hidden:        convert.BValOrFalse(raw[16]),
		Insure:        convert.BValOrFalse(raw[17]),
		Renew:         convert.BValOrFalse(raw[18]),
		RateReal:      convert.F64ValOrZero(raw[19]),
		NoClose:       convert.BValOrFalse(raw[20]),
		PositionPair:  convert.SValOrEmpty(raw[21]),
	}

	return
}

type HistoricalCredit Credit
type FundingCreditNew Credit
type FundingCreditUpdate Credit
type FundingCreditCancel Credit

type FundingCreditSnapshot struct {
	Snapshot []*Credit
}

func NewFundingCreditSnapshotFromRaw(raw []interface{}) (snap *FundingCreditSnapshot, err error) {
	if len(raw) == 0 {
		return
	}

	fcs := make([]*Credit, 0)
	switch raw[0].(type) {
	case []interface{}:
		for _, v := range raw {
			if l, ok := v.([]interface{}); ok {
				o, err := NewCreditFromRaw(l)
				if err != nil {
					return snap, err
				}
				fcs = append(fcs, o)
			}
		}
	default:
		return snap, fmt.Errorf("not a funding credit snapshot")
	}
	snap = &FundingCreditSnapshot{
		Snapshot: fcs,
	}

	return
}

type LoanStatus string

const (
	LoanStatusActive          LoanStatus = "ACTIVE"
	LoanStatusExecuted        LoanStatus = "EXECUTED"
	LoanStatusPartiallyFilled LoanStatus = "PARTIALLY FILLED"
	LoanStatusCanceled        LoanStatus = "CANCELED"
)

type Loan struct {
	ID            int64
	Symbol        string
	Side          string
	MTSCreated    int64
	MTSUpdated    int64
	Amount        float64
	Flags         interface{}
	Status        LoanStatus
	Rate          float64
	Period        int64
	MTSOpened     int64
	MTSLastPayout int64
	Notify        bool
	Hidden        bool
	Insure        bool
	Renew         bool
	RateReal      float64
	NoClose       bool
}

func NewLoanFromRaw(raw []interface{}) (o *Loan, err error) {
	if len(raw) < 21 {
		return o, fmt.Errorf("data slice too short (len=%d) for loan: %#v", len(raw), raw)
	}

	o = &Loan{
		ID:            convert.I64ValOrZero(raw[0]),
		Symbol:        convert.SValOrEmpty(raw[1]),
		Side:          convert.SValOrEmpty(raw[2]),
		MTSCreated:    convert.I64ValOrZero(raw[3]),
		MTSUpdated:    convert.I64ValOrZero(raw[4]),
		Amount:        convert.F64ValOrZero(raw[5]),
		Flags:         raw[6],
		Status:        LoanStatus(convert.SValOrEmpty(raw[7])),
		Rate:          convert.F64ValOrZero(raw[11]),
		Period:        convert.I64ValOrZero(raw[12]),
		MTSOpened:     convert.I64ValOrZero(raw[13]),
		MTSLastPayout: convert.I64ValOrZero(raw[14]),
		Notify:        convert.BValOrFalse(raw[15]),
		Hidden:        convert.BValOrFalse(raw[16]),
		Insure:        convert.BValOrFalse(raw[17]),
		Renew:         convert.BValOrFalse(raw[18]),
		RateReal:      convert.F64ValOrZero(raw[19]),
		NoClose:       convert.BValOrFalse(raw[20]),
	}

	return o, nil
}

type HistoricalLoan Loan
type FundingLoanNew Loan
type FundingLoanUpdate Loan
type FundingLoanCancel Loan

type FundingLoanSnapshot struct {
	Snapshot []*Loan
}

func NewFundingLoanSnapshotFromRaw(raw []interface{}) (snap *FundingLoanSnapshot, err error) {
	if len(raw) == 0 {
		return
	}

	fls := make([]*Loan, 0)
	switch raw[0].(type) {
	case []interface{}:
		for _, v := range raw {
			if l, ok := v.([]interface{}); ok {
				o, err := NewLoanFromRaw(l)
				if err != nil {
					return snap, err
				}
				fls = append(fls, o)
			}
		}
	default:
		return snap, fmt.Errorf("not a funding loan snapshot")
	}
	snap = &FundingLoanSnapshot{
		Snapshot: fls,
	}

	return
}

type FundingTrade struct {
	ID         int64
	Symbol     string
	MTSCreated int64
	OfferID    int64
	Amount     float64
	Rate       float64
	Period     int64
	Maker      int64
}

func NewFundingTradeFromRaw(raw []interface{}) (o *FundingTrade, err error) {
	if len(raw) < 8 {
		return o, fmt.Errorf("data slice too short for funding trade: %#v", raw)
	}

	o = &FundingTrade{
		ID:         convert.I64ValOrZero(raw[0]),
		Symbol:     convert.SValOrEmpty(raw[1]),
		MTSCreated: convert.I64ValOrZero(raw[2]),
		OfferID:    convert.I64ValOrZero(raw[3]),
		Amount:     convert.F64ValOrZero(raw[4]),
		Rate:       convert.F64ValOrZero(raw[5]),
		Period:     convert.I64ValOrZero(raw[6]),
		Maker:      convert.I64ValOrZero(raw[7]),
	}

	return
}

type FundingTradeExecution FundingTrade
type FundingTradeUpdate FundingTrade
type FundingTradeSnapshot struct {
	Snapshot []*FundingTrade
}
type HistoricalFundingTradeSnapshot FundingTradeSnapshot

func NewFundingTradeSnapshotFromRaw(raw []interface{}) (snap *FundingTradeSnapshot, err error) {
	if len(raw) == 0 {
		return
	}

	fts := make([]*FundingTrade, 0)
	switch raw[0].(type) {
	case []interface{}:
		for _, v := range raw {
			if l, ok := v.([]interface{}); ok {
				o, err := NewFundingTradeFromRaw(l)
				if err != nil {
					return snap, err
				}
				fts = append(fts, o)
			}
		}
	default:
		return snap, fmt.Errorf("not a funding trade snapshot")
	}
	snap = &FundingTradeSnapshot{
		Snapshot: fts,
	}

	return
}

type Notification struct {
	MTS        int64
	Type       string
	MessageID  int64
	NotifyInfo interface{}
	Code       int64
	Status     string
	Text       string
}

func NewNotificationFromRaw(raw []interface{}) (o *Notification, err error) {
	if len(raw) < 8 {
		return o, fmt.Errorf("data slice too short for notification: %#v", raw)
	}

	o = &Notification{
		MTS:       convert.I64ValOrZero(raw[0]),
		Type:      convert.SValOrEmpty(raw[1]),
		MessageID: convert.I64ValOrZero(raw[2]),
		//NotifyInfo: raw[4],
		Code:   convert.I64ValOrZero(raw[5]),
		Status: convert.SValOrEmpty(raw[6]),
		Text:   convert.SValOrEmpty(raw[7]),
	}

	// raw[4] = notify info
	var nraw []interface{}
	if raw[4] != nil {
		nraw = raw[4].([]interface{})
		switch o.Type {
		case "on-req":
			if len(nraw) <= 0 {
				o.NotifyInfo = nil
				break
			}
			// will be a set of orders if created via rest
			// this is to accommodate OCO orders
			if _, ok := nraw[0].([]interface{}); ok {
				o.NotifyInfo, err = NewOrderSnapshotFromRaw(nraw)
				if err != nil {
					return nil, err
				}
			} else {
				on, err := NewOrderFromRaw(nraw)
				if err != nil {
					return nil, err
				}
				oNew := OrderNew(*on)
				o.NotifyInfo = &oNew
			}
		case "ou-req":
			on, err := NewOrderFromRaw(nraw)
			if err != nil {
				return nil, err
			}
			oNew := OrderUpdate(*on)
			o.NotifyInfo = &oNew
		case "oc-req":
			// if list of list then parse to order snapshot
			oc, err := NewOrderFromRaw(nraw)
			if err != nil {
				return o, err
			}
			orderCancel := OrderCancel(*oc)
			o.NotifyInfo = &orderCancel
		case "fon-req":
			fon, err := NewOfferFromRaw(nraw)
			if err != nil {
				return o, err
			}
			fundingOffer := FundingOfferNew(*fon)
			o.NotifyInfo = &fundingOffer
		case "foc-req":
			foc, err := NewOfferFromRaw(nraw)
			if err != nil {
				return o, err
			}
			fundingOffer := FundingOfferCancel(*foc)
			o.NotifyInfo = &fundingOffer
		case "uca":
			o.NotifyInfo = raw[4]
		case "acc_tf":
			o.NotifyInfo = raw[4]
		case "pm-req":
			p, err := NewPositionFromRaw(nraw)
			if err != nil {
				return o, err
			}
			cp := PositionCancel(*p)
			o.NotifyInfo = &cp
		default:
			o.NotifyInfo = raw[4]
		}
	}

	return
}

type Ticker struct {
	Symbol          string
	Frr             float64
	Bid             float64
	BidPeriod       int64
	BidSize         float64
	Ask             float64
	AskPeriod       int64
	AskSize         float64
	DailyChange     float64
	DailyChangePerc float64
	LastPrice       float64
	Volume          float64
	High            float64
	Low             float64
}

type TickerUpdate Ticker
type TickerSnapshot struct {
	Snapshot []*Ticker
}

func NewTickerSnapshotFromRaw(symbol string, raw [][]float64) (*TickerSnapshot, error) {
	if len(raw) <= 0 {
		return nil, fmt.Errorf("data slice too short for ticker snapshot: %#v", raw)
	}
	snap := make([]*Ticker, 0)
	for _, f := range raw {
		c, err := NewTickerFromRaw(symbol, ToInterface(f))
		if err == nil {
			snap = append(snap, c)
		}
	}
	return &TickerSnapshot{Snapshot: snap}, nil
}

func NewTickerFromRaw(symbol string, raw []interface{}) (t *Ticker, err error) {
	if len(raw) < 10 {
		return t, fmt.Errorf("data slice too short for ticker, expected %d got %d: %#v", 10, len(raw), raw)
	}
	// funding currency ticker
	// ignore bid/ask period for now
	if len(raw) == 13 {
		t = &Ticker{
			Symbol:          symbol,
			Bid:             convert.F64ValOrZero(raw[1]),
			BidSize:         convert.F64ValOrZero(raw[2]),
			Ask:             convert.F64ValOrZero(raw[4]),
			AskSize:         convert.F64ValOrZero(raw[5]),
			DailyChange:     convert.F64ValOrZero(raw[7]),
			DailyChangePerc: convert.F64ValOrZero(raw[8]),
			LastPrice:       convert.F64ValOrZero(raw[9]),
			Volume:          convert.F64ValOrZero(raw[10]),
			High:            convert.F64ValOrZero(raw[11]),
			Low:             convert.F64ValOrZero(raw[12]),
		}
		return t, nil
	} else if len(raw) == 16 {
		// on funding currencies (ex. fUSD)
		// SYMBOL, FRR, BID, BID_PERIOD, BID_SIZE, ASK, ASK_PERIOD, ASK_SIZE, DAILY_CHANGE, DAILY_CHANGE_RELATIVE,
		// LAST_PRICE, VOLUME, HIGH, LOW, _PLACEHOLDER, _PLACEHOLDER, FRR_AMOUNT_AVAILABLE
		t = &Ticker{
			Symbol:          symbol,
			Frr:             convert.F64ValOrZero(raw[0]),
			Bid:             convert.F64ValOrZero(raw[1]),
			BidPeriod:       convert.I64ValOrZero(raw[2]),
			BidSize:         convert.F64ValOrZero(raw[3]),
			Ask:             convert.F64ValOrZero(raw[4]),
			AskPeriod:       convert.I64ValOrZero(raw[5]),
			AskSize:         convert.F64ValOrZero(raw[6]),
			DailyChange:     convert.F64ValOrZero(raw[7]),
			DailyChangePerc: convert.F64ValOrZero(raw[8]),
			LastPrice:       convert.F64ValOrZero(raw[9]),
			Volume:          convert.F64ValOrZero(raw[10]),
			High:            convert.F64ValOrZero(raw[11]),
			Low:             convert.F64ValOrZero(raw[12]),
		}
		return t, nil
	}

	// all other tickers
	// on trading pairs (ex. tBTCUSD)
	// SYMBOL, BID, BID_SIZE, ASK, ASK_SIZE, DAILY_CHANGE, DAILY_CHANGE_RELATIVE, LAST_PRICE, VOLUME, HIGH, LOW
	t = &Ticker{
		Symbol:          symbol,
		Bid:             convert.F64ValOrZero(raw[0]),
		BidSize:         convert.F64ValOrZero(raw[1]),
		Ask:             convert.F64ValOrZero(raw[2]),
		AskSize:         convert.F64ValOrZero(raw[3]),
		DailyChange:     convert.F64ValOrZero(raw[4]),
		DailyChangePerc: convert.F64ValOrZero(raw[5]),
		LastPrice:       convert.F64ValOrZero(raw[6]),
		Volume:          convert.F64ValOrZero(raw[7]),
		High:            convert.F64ValOrZero(raw[8]),
		Low:             convert.F64ValOrZero(raw[9]),
	}

	return t, nil
}

func NewTickerFromRestRaw(raw []interface{}) (t *Ticker, err error) {
	return NewTickerFromRaw(raw[0].(string), raw[1:])
}

type bookAction byte

// BookAction represents a new/update or removal for a book entry.
type BookAction bookAction

const (
	//BookUpdateEntry represents a new or updated book entry.
	BookUpdateEntry BookAction = 0
	//BookRemoveEntry represents a removal of a book entry.
	BookRemoveEntry BookAction = 1
)

// BookUpdate represents an order book price update.
type BookUpdate struct {
	ID          int64       // the book update ID, optional
	Symbol      string      // book symbol
	Price       float64     // updated price
	PriceJsNum  json.Number // update price as json.Number
	Count       int64       // updated count, optional
	Amount      float64     // updated amount
	AmountJsNum json.Number // update amount as json.Number
	Side        OrderSide   // side
	Action      BookAction  // action (add/remove)
}

type BookUpdateSnapshot struct {
	Snapshot []*BookUpdate
}

func NewBookUpdateSnapshotFromRaw(symbol, precision string, raw [][]float64, raw_numbers interface{}) (*BookUpdateSnapshot, error) {
	if len(raw) <= 0 {
		return nil, fmt.Errorf("data slice too short for book snapshot: %#v", raw)
	}
	snap := make([]*BookUpdate, len(raw))
	for i, f := range raw {
		b, err := NewBookUpdateFromRaw(symbol, precision, ToInterface(f), raw_numbers.([]interface{})[i])
		if err != nil {
			return nil, err
		}
		snap[i] = b
	}
	return &BookUpdateSnapshot{Snapshot: snap}, nil
}

func IsRawBook(precision string) bool {
	return precision == "R0"
}

// NewBookUpdateFromRaw creates a new book update object from raw data.  Precision determines how
// to interpret the side (baked into Count versus Amount)
// raw book updates [ID, price, qty], aggregated book updates [price, amount, count]
func NewBookUpdateFromRaw(symbol, precision string, data []interface{}, raw_numbers interface{}) (b *BookUpdate, err error) {
	if len(data) < 3 {
		return b, fmt.Errorf("data slice too short for book update, expected %d got %d: %#v", 5, len(data), data)
	}
	var px float64
	var px_num json.Number
	var id, cnt int64
	raw_num_array := raw_numbers.([]interface{})
	amt := convert.F64ValOrZero(data[2])
	amt_num := convert.FloatToJsonNumber(raw_num_array[2])

	var side OrderSide
	var actionCtrl float64
	if IsRawBook(precision) {
		// [ID, price, amount]
		id = convert.I64ValOrZero(data[0])
		px = convert.F64ValOrZero(data[1])
		px_num = convert.FloatToJsonNumber(raw_num_array[1])
		actionCtrl = px
	} else {
		// [price, amount, count]
		px = convert.F64ValOrZero(data[0])
		px_num = convert.FloatToJsonNumber(raw_num_array[0])
		cnt = convert.I64ValOrZero(data[1])
		actionCtrl = float64(cnt)
	}

	if amt > 0 {
		side = Bid
	} else {
		side = Ask
	}

	var action BookAction
	if actionCtrl <= 0 {
		action = BookRemoveEntry
	} else {
		action = BookUpdateEntry
	}

	b = &BookUpdate{
		Symbol:      symbol,
		Price:       math.Abs(px),
		PriceJsNum:  px_num,
		Count:       cnt,
		Amount:      math.Abs(amt),
		AmountJsNum: amt_num,
		Side:        side,
		Action:      action,
		ID:          id,
	}

	return
}

type Candle struct {
	Symbol     string
	Resolution CandleResolution
	MTS        int64
	Open       float64
	Close      float64
	High       float64
	Low        float64
	Volume     float64
}

type CandleSnapshot struct {
	Snapshot []*Candle
}

func ToFloat64Slice(slice []interface{}) []float64 {
	data := make([]float64, 0, len(slice))
	for _, i := range slice {
		if f, ok := i.(float64); ok {
			data = append(data, f)
		}
	}
	return data
}

func ToInterface(flt []float64) []interface{} {
	data := make([]interface{}, len(flt))
	for j, f := range flt {
		data[j] = f
	}
	return data
}

func NewCandleSnapshotFromRaw(symbol string, resolution CandleResolution, raw [][]float64) (*CandleSnapshot, error) {
	if len(raw) <= 0 {
		return nil, fmt.Errorf("data slice too short for candle snapshot: %#v", raw)
	}
	snap := make([]*Candle, 0)
	for _, f := range raw {
		c, err := NewCandleFromRaw(symbol, resolution, ToInterface(f))
		if err == nil {
			snap = append(snap, c)
		}
	}
	return &CandleSnapshot{Snapshot: snap}, nil
}

func NewCandleFromRaw(symbol string, resolution CandleResolution, raw []interface{}) (c *Candle, err error) {
	if len(raw) < 6 {
		return c, fmt.Errorf("data slice too short for candle, expected %d got %d: %#v", 6, len(raw), raw)
	}

	c = &Candle{
		Symbol:     symbol,
		Resolution: resolution,
		MTS:        convert.I64ValOrZero(raw[0]),
		Open:       convert.F64ValOrZero(raw[1]),
		Close:      convert.F64ValOrZero(raw[2]),
		High:       convert.F64ValOrZero(raw[3]),
		Low:        convert.F64ValOrZero(raw[4]),
		Volume:     convert.F64ValOrZero(raw[5]),
	}

	return
}

type Ledger struct {
	ID          int64
	Currency    string
	Nil1        float64
	MTS         int64
	Nil2        float64
	Amount      float64
	Balance     float64
	Nil3        float64
	Description string
}

// NewLedgerFromRaw takes the raw list of values as returned from the websocket
// service and tries to convert it into an Ledger.
func NewLedgerFromRaw(raw []interface{}) (o *Ledger, err error) {
	if len(raw) == 9 {
		o = &Ledger{
			ID:          int64(convert.F64ValOrZero(raw[0])),
			Currency:    convert.SValOrEmpty(raw[1]),
			Nil1:        convert.F64ValOrZero(raw[2]),
			MTS:         convert.I64ValOrZero(raw[3]),
			Nil2:        convert.F64ValOrZero(raw[4]),
			Amount:      convert.F64ValOrZero(raw[5]),
			Balance:     convert.F64ValOrZero(raw[6]),
			Nil3:        convert.F64ValOrZero(raw[7]),
			Description: convert.SValOrEmpty(raw[8]),
			// API returns 3 Nil values, what do they map to?
			// API documentation says ID is type integer but api returns a string
		}
	} else {
		return o, fmt.Errorf("data slice too short for ledger: %#v", raw)
	}
	return
}

type LedgerSnapshot struct {
	Snapshot []*Ledger
}

// LedgerSnapshotFromRaw takes a raw list of values as returned from the websocket
// service and tries to convert it into an LedgerSnapshot.
func NewLedgerSnapshotFromRaw(raw []interface{}) (s *LedgerSnapshot, err error) {
	if len(raw) == 0 {
		return s, fmt.Errorf("data slice too short for ledgers: %#v", raw)
	}

	os := make([]*Ledger, 0)
	switch raw[0].(type) {
	case []interface{}:
		for _, v := range raw {
			if l, ok := v.([]interface{}); ok {
				o, err := NewLedgerFromRaw(l)
				if err != nil {
					return s, err
				}
				os = append(os, o)
			}
		}
	default:
		return s, fmt.Errorf("not an ledger snapshot")
	}
	s = &LedgerSnapshot{Snapshot: os}
	return
}

type CurrencyConf struct {
	Currency  string
	Label     string
	Symbol    string
	Pairs     []string
	Pools     []string
	Explorers ExplorerConf
	Unit      string
}

type ExplorerConf struct {
	BaseUri        string
	AddressUri     string
	TransactionUri string
}

type CurrencyConfigMapping string

const (
	CurrencyLabelMap    CurrencyConfigMapping = "pub:map:currency:label"
	CurrencySymbolMap   CurrencyConfigMapping = "pub:map:currency:sym"
	CurrencyUnitMap     CurrencyConfigMapping = "pub:map:currency:unit"
	CurrencyExplorerMap CurrencyConfigMapping = "pub:map:currency:explorer"
	CurrencyExchangeMap CurrencyConfigMapping = "pub:list:pair:exchange"
)

type RawCurrencyConf struct {
	Mapping string
	Data    interface{}
}

func parseCurrencyLabelMap(config map[string]CurrencyConf, raw []interface{}) {
	for _, rawLabel := range raw {
		data := rawLabel.([]interface{})
		cur := data[0].(string)
		if val, ok := config[cur]; ok {
			// add value
			val.Label = data[1].(string)
			config[cur] = val
		} else {
			// create new empty config instance
			cfg := CurrencyConf{}
			cfg.Label = data[1].(string)
			cfg.Currency = cur
			config[cur] = cfg
		}
	}
}

func parseCurrencySymbMap(config map[string]CurrencyConf, raw []interface{}) {
	for _, rawLabel := range raw {
		data := rawLabel.([]interface{})
		cur := data[0].(string)
		if val, ok := config[cur]; ok {
			// add value
			val.Symbol = data[1].(string)
			config[cur] = val
		} else {
			// create new empty config instance
			cfg := CurrencyConf{}
			cfg.Symbol = data[1].(string)
			cfg.Currency = cur
			config[cur] = cfg
		}
	}
}

func parseCurrencyUnitMap(config map[string]CurrencyConf, raw []interface{}) {
	for _, rawLabel := range raw {
		data := rawLabel.([]interface{})
		cur := data[0].(string)
		if val, ok := config[cur]; ok {
			// add value
			val.Unit = data[1].(string)
			config[cur] = val
		} else {
			// create new empty config instance
			cfg := CurrencyConf{}
			cfg.Unit = data[1].(string)
			cfg.Currency = cur
			config[cur] = cfg
		}
	}
}

func parseCurrencyExplorerMap(config map[string]CurrencyConf, raw []interface{}) {
	for _, rawLabel := range raw {
		data := rawLabel.([]interface{})
		cur := data[0].(string)
		explorers := data[1].([]interface{})
		var cfg CurrencyConf
		if val, ok := config[cur]; ok {
			cfg = val
		} else {
			// create new empty config instance
			cc := CurrencyConf{}
			cc.Currency = cur
			cfg = cc
		}
		ec := ExplorerConf{
			explorers[0].(string),
			explorers[1].(string),
			explorers[2].(string),
		}
		cfg.Explorers = ec
		config[cur] = cfg
	}
}

func parseCurrencyExchangeMap(config map[string]CurrencyConf, raw []interface{}) {
	for _, rs := range raw {
		symbol := rs.(string)
		var base, quote string
		if len(symbol) > 6 {
			base = strings.Split(symbol, ":")[0]
			quote = strings.Split(symbol, ":")[1]
		} else {
			base = symbol[3:]
			quote = symbol[:3]
		}
		// append if base exists in configs
		if val, ok := config[base]; ok {
			val.Pairs = append(val.Pairs, symbol)
			config[base] = val
		}
		// append if quote exists in configs
		if val, ok := config[quote]; ok {
			val.Pairs = append(val.Pairs, symbol)
			config[quote] = val
		}
	}
}

func NewCurrencyConfFromRaw(raw []RawCurrencyConf) ([]CurrencyConf, error) {
	configMap := make(map[string]CurrencyConf)
	for _, r := range raw {
		switch CurrencyConfigMapping(r.Mapping) {
		case CurrencyLabelMap:
			data := r.Data.([]interface{})
			parseCurrencyLabelMap(configMap, data)
		case CurrencySymbolMap:
			data := r.Data.([]interface{})
			parseCurrencySymbMap(configMap, data)
		case CurrencyUnitMap:
			data := r.Data.([]interface{})
			parseCurrencyUnitMap(configMap, data)
		case CurrencyExplorerMap:
			data := r.Data.([]interface{})
			parseCurrencyExplorerMap(configMap, data)
		case CurrencyExchangeMap:
			data := r.Data.([]interface{})
			parseCurrencyExchangeMap(configMap, data)
		}
	}
	// convert map to array
	configs := make([]CurrencyConf, 0)
	for _, v := range configMap {
		configs = append(configs, v)
	}
	return configs, nil
}

type StatKey string

const (
	FundingSizeKey   StatKey = "funding.size"
	CreditSizeKey    StatKey = "credits.size"
	CreditSizeSymKey StatKey = "credits.size.sym"
	PositionSizeKey  StatKey = "pos.size"
)

type Stat struct {
	Period int64
	Volume float64
}

type DerivativeStatusSnapshot struct {
	Snapshot []*DerivativeStatus
}

type StatusType string

const (
	DerivativeStatusType StatusType = "deriv"
)

type DerivativeStatus struct {
	Symbol               string
	MTS                  int64
	Price                float64
	SpotPrice            float64
	InsuranceFundBalance float64
	FundingAccrued       float64
	FundingStep          float64
}

func NewDerivativeStatusFromWsRaw(symbol string, raw []interface{}) (*DerivativeStatus, error) {
	if len(raw) == 11 {
		ds := &DerivativeStatus{
			Symbol: symbol,
			MTS:    convert.I64ValOrZero(raw[0]),
			// placeholder
			Price:     convert.F64ValOrZero(raw[2]),
			SpotPrice: convert.F64ValOrZero(raw[3]),
			// placeholder
			InsuranceFundBalance: convert.F64ValOrZero(raw[5]),
			// placeholder
			// placeholder
			FundingAccrued: convert.F64ValOrZero(raw[8]),
			FundingStep:    convert.F64ValOrZero(raw[9]),
			// placeholder
		}
		return ds, nil
	} else {
		return nil, fmt.Errorf("data slice too short for derivative status: %#v", raw)
	}
}

func NewDerivativeStatusFromRaw(raw []interface{}) (*DerivativeStatus, error) {
	if len(raw) == 12 {
		ds := &DerivativeStatus{
			Symbol: convert.SValOrEmpty(raw[0]),
			MTS:    convert.I64ValOrZero(raw[1]),
			// placeholder
			Price:     convert.F64ValOrZero(raw[3]),
			SpotPrice: convert.F64ValOrZero(raw[4]),
			// placeholder
			InsuranceFundBalance: convert.F64ValOrZero(raw[6]),
			// placeholder
			// placeholder
			FundingAccrued: convert.F64ValOrZero(raw[9]),
			FundingStep:    convert.F64ValOrZero(raw[10]),
			// placeholder
		}
		return ds, nil
	} else {
		return nil, fmt.Errorf("data slice too short for derivative status: %#v", raw)
	}
}

func NewDerivativeSnapshotFromRaw(raw [][]interface{}) (*DerivativeStatusSnapshot, error) {
	snapshot := make([]*DerivativeStatus, len(raw))
	for i, rStatus := range raw {
		pStatus, err := NewDerivativeStatusFromRaw(rStatus)
		if err != nil {
			return nil, err
		}
		snapshot[i] = pStatus
	}
	return &DerivativeStatusSnapshot{Snapshot: snapshot}, nil
}
