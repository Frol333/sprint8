package main

import (
	"database/sql"
	"fmt"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	// реализуйте добавление строки в таблицу parcel, используйте данные из переменной p
	res, err := s.db.Exec("INSERT INTO parcel (client, status, address, created_at) VALUES (?, ?, ?, ?)",
		p.Client, p.Status, p.Address, p.CreatedAt)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil

}

func (s ParcelStore) Get(number int) (Parcel, error) {
	// реализуйте чтение строки по заданному number
	// здесь из таблицы должна вернуться только одна строка
	row := s.db.QueryRow("SELECT number, client, status, address, created_at FROM parcel WHERE number = ?", number)
	p := Parcel{}
	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		return Parcel{}, err
	}
	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	// Выполняем запрос к базе данных для получения посылок по заданному client.
	rows, err := s.db.Query("SELECT number, client, status, address, created_at FROM parcel WHERE client = ?", client)
	if err != nil {
		// Если запрос не удался, возвращаем ошибку.
		return nil, err
	}
	defer rows.Close()

	var res []Parcel // Инициализируем слайс для хранения результатов.

	// Итерируемся по каждой строке результата запроса.
	for rows.Next() {
		p := Parcel{} // Создаем новый экземпляр структуры Parcel для каждой строки.
		// Сканируем значения из текущей строки в поля структуры Parcel.
		err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
		if err != nil {
			// Если произошла ошибка при сканировании, возвращаем ее.
			return nil, err
		}
		res = append(res, p) // Добавляем заполненную структуру Parcel в слайс результатов.
	}

	// Проверяем наличие ошибок после завершения итерации по строкам.
	if err := rows.Err(); err != nil {
		return nil, err // Возвращаем ошибку, если она возникла при итерации.
	}

	return res, nil // Возвращаем слайс с результатами и nil в качестве ошибки (если все прошло успешно).
}

func (s ParcelStore) SetStatus(number int, status string) error {
	// реализуйте обновление статуса в таблице parcel
	_, err := s.db.Exec("UPDATE parcel SET status = ? WHERE number = ?", status, number)
	return err

}

func (s ParcelStore) SetAddress(number int, address string) error {
	// Реализуем обновление адреса посылки в таблице parcel.
	// Адрес можно менять только если статус посылки 'registered'.

	// Используем транзакцию, чтобы обеспечить атомарность операции.
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // Откатываем транзакцию в случае ошибки.

	// Выполняем UPDATE с условием, чтобы изменить адрес только если статус registered.
	result, err := tx.Exec("UPDATE parcel SET address = ? WHERE number = ? AND status = ?", address, number, ParcelStatusRegistered)
	if err != nil {
		return err
	}

	// Проверяем, была ли изменена хотя бы одна строка.
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		// Если ни одна строка не была изменена, значит, либо посылки с таким номером не существует,
		// либо ее статус не 'registered'.
		// В данном случае возвращаем ошибку, так как нет возможности узнать наверняка, поэтому просто говорим, что изменение не удалось.
		return fmt.Errorf("не удалось изменить адрес посылки")
	}

	// Фиксируем транзакцию.
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (s ParcelStore) Delete(number int) error {
	// Пытаемся удалить строку из таблицы parcel по заданному номеру.
	// Удаление разрешено, только если статус посылки 'registered'.
	// Используем один UPDATE запрос, чтобы избежать дополнительного SELECT.

	result, err := s.db.Exec("DELETE FROM parcel WHERE number = ? AND status = ?", number, ParcelStatusRegistered)
	if err != nil {
		// Если произошла ошибка при выполнении запроса, возвращаем её.
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		// Если не удалось получить количество затронутых строк, возвращаем ошибку.
		return err
	}

	if rowsAffected == 0 {
		// Если ни одна строка не была удалена, значит, либо посылки с таким номером не существует,
		// либо её статус не 'registered'. В таком случае возвращаем ошибку.
		return fmt.Errorf("удаление не удалось: либо посылки с таким номером не существует, либо статус не 'registered'")
	}

	// Если строка была успешно удалена, возвращаем nil (нет ошибки).
	return nil
}
