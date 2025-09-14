package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки

// TestSetAddress проверяет обновление адреса
func TestAddGetDelete(t *testing.T) {
	db, err := sql.Open("sqlite", "test.db") // connection
	require.NoError(t, err)
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel) // add
	require.NoError(t, err)
	require.NotEmpty(t, id)

	storedParcel, err := store.Get(id) // get
	require.NoError(t, err)
	require.Equal(t, parcel.Address, storedParcel.Address)

	err = store.Delete(id) // delete
	require.NoError(t, err)

	_, err = store.Get(id)
	require.Error(t, err)
}

// ... (TestSetAddress, TestSetStatus, TestGetByClient - аналогично: connect, add, set/get, check, delete)

func TestSetAddress(t *testing.T) {
	db, err := sql.Open("sqlite", "test.db") // connection
	require.NoError(t, err)
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel) // add
	require.NoError(t, err)
	require.NotEmpty(t, id)
	newAddress := "new test address"

	err = store.SetAddress(id, newAddress) // set address
	require.NoError(t, err)

	storedParcel, err := store.Get(id) // check
	require.NoError(t, err)
	require.Equal(t, newAddress, storedParcel.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	db, err := sql.Open("sqlite", "test.db") // connection
	require.NoError(t, err)
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel) // add
	require.NoError(t, err)
	require.NotEmpty(t, id)

	err = store.SetStatus(id, ParcelStatusSent) // set status
	require.NoError(t, err)

	storedParcel, err := store.Get(id) // check
	require.NoError(t, err)
	require.Equal(t, ParcelStatusSent, storedParcel.Status)
}

func TestGetByClient(t *testing.T) {
	db, err := sql.Open("sqlite", "test.db") // connection
	require.NoError(t, err)
	defer db.Close()
	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	client := randRange.Intn(10_000_000)
	for i := range parcels {
		parcels[i].Client = client
	}
	for i := range parcels {
		id, err := store.Add(parcels[i]) // add
		require.NoError(t, err)
		parcels[i].Number = id
		parcelMap[id] = parcels[i]
	}

	storedParcels, err := store.GetByClient(client) // get by client
	require.NoError(t, err)
	require.Len(t, storedParcels, len(parcels))

	for _, parcel := range storedParcels { // check
		expectedParcel, ok := parcelMap[parcel.Number]
		require.True(t, ok, "Parcel not found in map")
		require.Equal(t, expectedParcel.Client, parcel.Client)
		require.Equal(t, expectedParcel.Status, parcel.Status)

	}
}
