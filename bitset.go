package reflector

import (
	"encoding/json"
	"fmt"
	"sort"
)

type (
	BitIndex   uint64
	BitIndexes []BitIndex
	BitSet     []byte
	UserRole   struct {
		Rights     string
		Permisions BitSet
	}
)

const (
	Login BitIndex = 1

	DashboardCompany  BitIndex = 5
	DashboardPersonal BitIndex = 6

	UserList   BitIndex = 8
	UserAdd    BitIndex = 9
	UserEdit   BitIndex = 10
	UserDelete BitIndex = 11

	RoleList                 BitIndex = 16
	RoleAdd                  BitIndex = 17
	RoleEdit                 BitIndex = 18
	RoleDelete               BitIndex = 19
	ListSchedulesForAllUsers BitIndex = 20

	InventoryList   BitIndex = 24
	InventoryAdd    BitIndex = 25
	InventoryEdit   BitIndex = 26
	InventoryDelete BitIndex = 27

	QuoteList   BitIndex = 32
	QuoteAdd    BitIndex = 33
	QuoteEdit   BitIndex = 34
	QuoteDelete BitIndex = 35
	QuoteAssign BitIndex = 36

	QuotesAssignedList BitIndex = 37
	QuotesAssignedAdd  BitIndex = 38
	QuotesAssignedEdit          = 39

	ContactList    BitIndex = 40
	ContactAdd     BitIndex = 41
	ContactEdit    BitIndex = 42
	ContactDelete  BitIndex = 43
	ContactListAll BitIndex = 44

	OrderList          BitIndex = 48
	OrderAdd           BitIndex = 49
	OrderEdit          BitIndex = 50
	OrderDelete        BitIndex = 51
	OrderAssign        BitIndex = 52
	OrdersAssignedList BitIndex = 53
	OrdersAssignedAdd  BitIndex = 54
	OrdersAssignedEdit BitIndex = 55

	CustomerList   BitIndex = 56
	CustomerAdd    BitIndex = 57
	CustomerEdit   BitIndex = 58
	CustomerDelete BitIndex = 59

	AddressList    BitIndex = 64
	AddressAdd     BitIndex = 65
	AddressEdit    BitIndex = 66
	AddressDelete  BitIndex = 67
	AddressListAll BitIndex = 68

	GreetingsList         BitIndex = 160
	GreetingsAdd          BitIndex = 161
	GreetingsEdit         BitIndex = 162
	GreetingsDelete       BitIndex = 163
	GreetingsAssignedList BitIndex = 164
	GreetingsAssignedAdd  BitIndex = 165
	GreetingsAssignedEdit BitIndex = 166

	ShippingList   BitIndex = 178
	ShippingAdd    BitIndex = 179
	ShippingEdit   BitIndex = 180
	ShippingDelete BitIndex = 181

	InstallationList   BitIndex = 185
	InstallationAdd    BitIndex = 186
	InstallationEdit   BitIndex = 187
	InstallationDelete BitIndex = 188

	ReportsAdd    BitIndex = 192
	ReportsList   BitIndex = 193
	ReportsEdit   BitIndex = 194
	ReportsDelete BitIndex = 195

	CalendarAdd    BitIndex = 198
	CalendarList   BitIndex = 199
	CalendarEdit   BitIndex = 200
	CalendarDelete BitIndex = 201

	SettingsGroupAdd           BitIndex = 202
	SettingsGroupList          BitIndex = 203
	SettingsGroupEdit          BitIndex = 204
	SettingsGroupDelete        BitIndex = 205
	SettingsVoicesAdd          BitIndex = 206
	SettingsVoicesList         BitIndex = 207
	SettingsVoicesEdit         BitIndex = 208
	SettingsVoicesDelete       BitIndex = 209
	SettingsItemsAdd           BitIndex = 210
	SettingsItemsList          BitIndex = 211
	SettingsItemsEdit          BitIndex = 212
	SettingsItemsDelete        BitIndex = 213
	SettingsTaxesAdd           BitIndex = 214
	SettingsTaxesList          BitIndex = 215
	SettingsTaxesEdit          BitIndex = 216
	SettingsTaxesDelete        BitIndex = 217
	SettingsTemplatesAdd       BitIndex = 218
	SettingsTemplatesList      BitIndex = 219
	SettingsTemplatesEdit      BitIndex = 220
	SettingsTemplatesDelete    BitIndex = 221
	SettingsIntegrationAdd     BitIndex = 222
	SettingsIntegrationList    BitIndex = 223
	SettingsIntegrationEdit    BitIndex = 224
	SettingsIntegrationDelete  BitIndex = 225
	SettingsAlertsAdd          BitIndex = 226
	SettingsAlertsList         BitIndex = 227
	SettingsAlertsEdit         BitIndex = 228
	SettingsAlertsDelete       BitIndex = 229
	TasksAdd                   BitIndex = 230
	TasksList                  BitIndex = 231
	TasksEdit                  BitIndex = 232
	TasksDelete                BitIndex = 233
	TasksAssignedAdd           BitIndex = 234
	TasksAssignedList          BitIndex = 235
	TasksAssignedEdit          BitIndex = 236
	TasksAssignedDelete        BitIndex = 237
	SettingsGroupedItemsAdd    BitIndex = 238
	SettingsGroupedItemsList   BitIndex = 239
	SettingsGroupedItemsEdit   BitIndex = 240
	SettingsGroupedItemsDelete BitIndex = 241
	SettingsAutomationsAdd     BitIndex = 242
	SettingsAutomationsList    BitIndex = 243
	SettingsAutomationsEdit    BitIndex = 244
	SettingsAutomationsDelete  BitIndex = 245
	SettingsCommissionsAdd     BitIndex = 246
	SettingsCommissionsList    BitIndex = 247
	SettingsCommissionsEdit    BitIndex = 248
	SettingsCommissionsDelete  BitIndex = 249
	SettingsCarriersAdd        BitIndex = 250
	SettingsCarriersList       BitIndex = 251
	SettingsCarriersEdit       BitIndex = 252
	SettingsCarriersDelete     BitIndex = 253

	RoutesList BitIndex = 258
	RoutesAdd  BitIndex = 259
	RoutesEdit BitIndex = 260
)

var (
	bitIndexesMap = map[BitIndex]string{
		Login:                      "customer.login",
		UserList:                   "users.list",
		UserAdd:                    "users.add",
		UserEdit:                   "users.edit",
		UserDelete:                 "users.delete",
		RoleList:                   "userroles.list",
		RoleAdd:                    "userroles.add",
		RoleEdit:                   "userroles.edit",
		RoleDelete:                 "userroles.delete",
		ListSchedulesForAllUsers:   "userschedule.list.all",
		InventoryList:              "inventory.list",
		InventoryAdd:               "inventory.add",
		InventoryEdit:              "inventory.edit",
		InventoryDelete:            "inventory.delete",
		QuoteList:                  "quotes.list",
		QuoteAdd:                   "quotes.create",
		QuoteEdit:                  "quotes.edit",
		QuoteDelete:                "quotes.delete",
		QuoteAssign:                "quotes.assign",
		ContactList:                "contacts.list.own",
		ContactAdd:                 "contacts.add",
		ContactEdit:                "contacts.edit",
		ContactDelete:              "contacts.delete",
		ContactListAll:             "contacts.list.all",
		OrderList:                  "orders.list",
		OrderAdd:                   "orders.add",
		OrderEdit:                  "orders.edit",
		OrderDelete:                "orders.delete",
		OrderAssign:                "orders.assign",
		CustomerList:               "customers.list",
		CustomerAdd:                "customers.add",
		CustomerEdit:               "customers.edit",
		CustomerDelete:             "customers.delete",
		AddressList:                "addresses.list.own",
		AddressAdd:                 "addresses.add",
		AddressEdit:                "addresses.edit",
		AddressDelete:              "addresses.delete",
		AddressListAll:             "addresses.list.all",
		DashboardCompany:           "dashboard.company",
		DashboardPersonal:          "dashboard.personal",
		QuotesAssignedList:         "quotes.assigned.list",
		QuotesAssignedAdd:          "quotes.assigned.add",
		QuotesAssignedEdit:         "quotes.assigned.edit",
		OrdersAssignedList:         "orders.assigned.list",
		OrdersAssignedAdd:          "orders.assigned.add",
		OrdersAssignedEdit:         "orders.assigned.edit",
		GreetingsList:              "suborders.greetings.list",
		GreetingsAdd:               "suborders.greetings.add",
		GreetingsEdit:              "suborders.greetings.edit",
		GreetingsDelete:            "suborders.greetings.delete",
		GreetingsAssignedList:      "suborders.greetings.assigned.list",
		GreetingsAssignedAdd:       "suborders.greetings.assigned.add",
		GreetingsAssignedEdit:      "suborders.greetings.assigned.edit",
		ShippingList:               "suborders.shipping.list",
		ShippingAdd:                "suborders.shipping.add",
		ShippingEdit:               "suborders.shipping.edit",
		ShippingDelete:             "suborders.shipping.delete",
		InstallationList:           "suborders.installations.list",
		InstallationAdd:            "suborders.installations.add",
		InstallationEdit:           "suborders.installations.edit",
		InstallationDelete:         "suborders.installations.delete",
		ReportsAdd:                 "reports.add",
		ReportsList:                "reports.list",
		ReportsEdit:                "reports.edit",
		ReportsDelete:              "reports.delete",
		CalendarAdd:                "calendar.add",
		CalendarList:               "calendar.list",
		CalendarEdit:               "calendar.edit",
		CalendarDelete:             "calendar.delete",
		SettingsGroupAdd:           "settings.groups.add",
		SettingsGroupList:          "settings.groups.list",
		SettingsGroupEdit:          "settings.groups.edit",
		SettingsGroupDelete:        "settings.groups.delete",
		SettingsVoicesAdd:          "settings.voices.add",
		SettingsVoicesList:         "settings.voices.list",
		SettingsVoicesEdit:         "settings.voices.edit",
		SettingsVoicesDelete:       "settings.voices.delete",
		SettingsItemsAdd:           "settings.items.add",
		SettingsItemsList:          "settings.items.list",
		SettingsItemsEdit:          "settings.items.edit",
		SettingsItemsDelete:        "settings.items.delete",
		SettingsTaxesAdd:           "settings.taxes.add",
		SettingsTaxesList:          "settings.taxes.list",
		SettingsTaxesEdit:          "settings.taxes.edit",
		SettingsTaxesDelete:        "settings.taxes.delete",
		SettingsTemplatesAdd:       "settings.templates.add",
		SettingsTemplatesList:      "settings.templates.list",
		SettingsTemplatesEdit:      "settings.templates.edit",
		SettingsTemplatesDelete:    "settings.templates.delete",
		SettingsIntegrationAdd:     "settings.integrations.add",
		SettingsIntegrationList:    "settings.integrations.list",
		SettingsIntegrationEdit:    "settings.integrations.edit",
		SettingsIntegrationDelete:  "settings.integrations.delete",
		SettingsAlertsAdd:          "settings.alerts.add",
		SettingsAlertsList:         "settings.alerts.list",
		SettingsAlertsEdit:         "settings.alerts.edit",
		SettingsAlertsDelete:       "settings.alerts.delete",
		SettingsGroupedItemsAdd:    "settings.groupedItems.add",
		SettingsGroupedItemsList:   "settings.groupedItems.list",
		SettingsGroupedItemsEdit:   "settings.groupedItems.edit",
		SettingsGroupedItemsDelete: "settings.groupedItems.delete",
		SettingsAutomationsAdd:     "settings.automation.add",
		SettingsAutomationsList:    "settings.automation.list",
		SettingsAutomationsEdit:    "settings.automation.edit",
		SettingsAutomationsDelete:  "settings.automation.delete",
		SettingsCommissionsAdd:     "settings.commissions.add",
		SettingsCommissionsList:    "settings.commissions.list",
		SettingsCommissionsEdit:    "settings.commissions.edit",
		SettingsCommissionsDelete:  "settings.commissions.delete",
		SettingsCarriersAdd:        "settings.carriers.add",
		SettingsCarriersList:       "settings.carriers.list",
		SettingsCarriersEdit:       "settings.carriers.edit",
		SettingsCarriersDelete:     "settings.carriers.delete",
		TasksAdd:                   "tasks.add",
		TasksList:                  "tasks.list",
		TasksEdit:                  "tasks.edit",
		TasksDelete:                "tasks.delete",
		TasksAssignedAdd:           "tasks.assigned.add",
		TasksAssignedList:          "tasks.assigned.list",
		TasksAssignedEdit:          "tasks.assigned.edit",
		TasksAssignedDelete:        "tasks.assigned.delete",

		RoutesList: "settings.routes.list",
		RoutesAdd:  "settings.routes.add",
		RoutesEdit: "settings.routes.edit",
	}

	allBitsIndexes       BitIndexes
	reverseBitIndexesMap = map[string]BitIndex{}
	neededBytes          int
)

func init() {
	// performing init
	neededBytes = int(0)
	for index, description := range bitIndexesMap {
		// searching the maximum number of bytes needed for this slice
		if int(index) > neededBytes {
			neededBytes = int(index)
		}
		// filling up the reverse map
		reverseBitIndexesMap[description] = index
		// filling up the slice with all indexes
		allBitsIndexes = append(allBitsIndexes, index)
	}
	// by default, rights are sorted - see implementation of Sorter interface below
	sort.Sort(allBitsIndexes)
	// set the needed number of bytes to perform new, load and make from bytearray
	neededBytes = (neededBytes >> 3) + 1
}

func New() BitSet {
	return make([]byte, neededBytes)
}

func MakeFromByteArray(bytes []byte) (BitSet, error) {
	s := make([]byte, neededBytes)
	if neededBytes != len(bytes) {
		return s, fmt.Errorf("Make from []byte Error : trying to load %d byte size into %d size.", len(bytes), neededBytes)
	}
	copy(s, bytes)
	return s, nil
}

func (bs *BitSet) Load(bytes []byte) error {
	if neededBytes != len(bytes) {
		return fmt.Errorf("Load Error : trying to load %d byte size into %d size.", len(bytes), neededBytes)
	}
	*bs = make([]byte, neededBytes)
	copy(*bs, bytes)
	return nil
}

func (bs BitSet) getByteAndBitNo(index BitIndex) (uint64, uint64) {
	byteNo := uint64(index >> 3)
	bitNo := uint64(index) - byteNo*8
	return byteNo, bitNo
}

func (bs BitSet) Set(index BitIndex) error {
	byteNo, bitNo := bs.getByteAndBitNo(index)
	if int(byteNo) > len(bs) {
		return fmt.Errorf("Error trying to set %d out of %d", byteNo, len(bs))
	}
	bs[byteNo] |= 1 << bitNo
	return nil
}

func (bs BitSet) Unset(index BitIndex) error {
	byteNo, bitNo := bs.getByteAndBitNo(index)
	if int(byteNo) > len(bs) {
		return fmt.Errorf("Error trying to set %d out of %d", byteNo, len(bs))
	}
	bs[byteNo] &= ^(1 << bitNo)
	return nil
}

func (bs BitSet) SetAll(perms ...BitIndex) {
	for _, perm := range perms {
		bs.Set(perm)
	}
}
func (bs BitSet) UnsetAll(perms ...BitIndex) {
	for _, perm := range perms {
		bs.Unset(perm)
	}
}

func (bs BitSet) Test(index BitIndex) (bool, error) {
	byteNo, bitNo := bs.getByteAndBitNo(index)
	if int(byteNo) > len(bs) {
		return false, fmt.Errorf("Error trying to set %d out of %d", byteNo, len(bs))

	}
	return bs[byteNo]&(1<<bitNo) != 0, nil
}

func (s BitSet) Len() int {
	return len(s)
}

func (bs BitSet) Bytes() []byte {
	return bs
}

//TODO : rename this
func (bs BitSet) Rights() ([]BitIndex, error) {
	var rights []BitIndex
	for _, right := range allBitsIndexes {
		ok, err := bs.Test(right)
		if err != nil {
			return nil, err
		}
		if ok {
			rights = append(rights, right)
		}
	}
	return rights, nil
}

//implementation of Stringer
func (bs BitSet) String() string {
	s := "User Rights :\n"
	for _, right := range allBitsIndexes {
		ok, err := bs.Test(right)
		if err != nil {
			return fmt.Sprintf("Error : %v", err)
		}
		if ok {
			s += right.String() + "\n"
		}
	}
	return s
}

func (i BitIndex) Name() string {
	return bitIndexesMap[i]
}

//implementation of Stringer
func (i BitIndex) String() string {
	return fmt.Sprintf("{%q:%d,%q:%q}", "bit", i, "description", bitIndexesMap[i])
}

// implementation of MarshalJSON (marshaller) for Right
func (i BitIndex) MarshalJSON() ([]byte, error) {
	return json.Marshal(bitIndexesMap[i])
}

//implementation of Sort for RightsSlice
func (s BitIndexes) Len() int {
	return len(s)
}

//implementation of Sort for RightsSlice
func (s BitIndexes) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

//implementation of Sort for RightsSlice
func (s BitIndexes) Less(i, j int) bool {
	return s[i] < s[j]
}

//TODO : rename this
func Rights() BitIndexes {
	return allBitsIndexes
}

//TODO : rename this
func RightsMap() map[BitIndex]string {
	return bitIndexesMap
}

//TODO : rename this
func ReverseMap() map[string]BitIndex {
	return reverseBitIndexesMap
}
