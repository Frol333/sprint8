package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

func TestAddGetDelete(t *testing.T) {
	db, err := sql.Open("sqlite", "test.db")
	require.NoError(t, err, "Ошибка при открытии соединения с БД") // require здесь уместен - тест не может продолжаться без соединения
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err, "Ошибка при добавлении посылки") // require здесь уместен - тест не может продолжаться без добавления
	require.NotEmpty(t, id, "ID не должен быть пустым")      // require здесь уместен - тест не может продолжаться без ID

	storedParcel, err := store.Get(id)
	require.NoError(t, err, "Ошибка при получении посылки") // require здесь уместен - тест не может продолжаться без получения

	assert.Equal(t, parcel.Client, storedParcel.Client, "Client не совпадает")
	assert.Equal(t, parcel.Status, storedParcel.Status, "Status не совпадает")
	assert.Equal(t, parcel.Address, storedParcel.Address, "Address не совпадает")

	err = store.Delete(id)
	require.NoError(t, err, "Ошибка при удалении посылки") // require здесь уместен - удаление необходимо для валидации последующего Get

	_, err = store.Get(id)
	require.Error(t, err, "Ожидалась ошибка при получении удаленной посылки") // require здесь уместен - тест проверяет ошибку
}

func TestSetAddress(t *testing.T) {
	db, err := sql.Open("sqlite", "test.db")
	require.NoError(t, err, "Ошибка при подключении к базе данных")
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err, "Ошибка при добавлении посылки")
	require.NotEmpty(t, id, "ID не должен быть пустым")

	newAddress := "new test address"

	err = store.SetAddress(id, newAddress)
	require.NoError(t, err, "Ошибка при обновлении адреса")

	storedParcel, err := store.Get(id)
	require.NoError(t, err, "Ошибка при получении посылки после обновления адреса")

	assert.Equal(t, newAddress, storedParcel.Address, "Адрес не был обновлен")
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	db, err := sql.Open("sqlite", "test.db")
	require.NoError(t, err, "Не удалось подключиться к базе данных. Тест не может быть продолжен.")
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err, "Не удалось добавить посылку. Тест не может быть продолжен.")
	require.NotEmpty(t, id, "ID посылки не должен быть пустым. Тест не может быть продолжен.")

	err = store.SetStatus(id, ParcelStatusSent)
	require.NoError(t, err, "Не удалось обновить статус посылки. Тест не может быть продолжен.")

	storedParcel, err := store.Get(id)
	require.NoError(t, err, "Не удалось получить посылку после обновления статуса. Тест не может быть продолжен.")

	assert.Equal(t, ParcelStatusSent, storedParcel.Status, "Статус посылки не соответствует ожидаемому.")
}

func TestGetByClient(t *testing.T) {
	db, err := sql.Open("sqlite", "test.db")
	require.NoError(t, err, "Ошибка подключения к базе данных")
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
		id, err := store.Add(parcels[i])
		require.NoError(t, err, "Ошибка при добавлении посылки")
		parcels[i].Number = id
		parcelMap[id] = parcels[i]
	}

	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err, "Ошибка при получении посылок по клиенту")

	// Тут assert.Len, так как тест может продолжить работу, даже если количество не совпадает
	assert.Len(t, storedParcels, len(parcels), "Неверное количество посылок")

	for _, parcel := range storedParcels {
		expectedParcel, ok := parcelMap[parcel.Number]
		require.True(t, ok, "Посылка не найдена в мапе") //критично, если в мапе нет, тест дальше не имеет смысла
		assert.Equal(t, expectedParcel.Client, parcel.Client, "Клиент не совпадает")
		assert.Equal(t, expectedParcel.Status, parcel.Status, "Статус не совпадает")

	}
}
