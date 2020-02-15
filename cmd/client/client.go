package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/akram8008/Bank-core/core"
	"log"
	"os"
	"strings"
)

func main() {
	// os.Stdin, os.Stout, os.Stderr, File
	file, err := os.OpenFile("log.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(file)
	log.Print("start application")
	log.Print("open db")
	db, err := sql.Open("sqlite3", "db.sqlite")
	if err != nil {
		log.Fatalf("can't open db: %v", err)
	}
	defer func() {
		log.Print("close db")
		if err := db.Close(); err != nil {
			log.Fatalf("can't close db: %v", err)
		}
	}()
	err = core.Init(db)
	if err != nil {
		log.Fatalf("can't init db: %v", err)
	}
	fmt.Fprintln(os.Stdout, "Добро пожаловать в наше приложение")
	log.Print("start operations loop")
	operationsLoop(db, "", unauthorizedOperations, unauthorizedOperationsLoop)
	log.Print("finish operations loop")
	log.Print("finish application")
}

func operationsLoop(db *sql.DB, login string, commands string, loop func(db *sql.DB, log string, cmd string) bool) {
	for {
		fmt.Println(commands)
		var cmd string
		_, err := fmt.Scan(&cmd)
		if err != nil {
			log.Fatalf("Can't read input: %v", err) // %v - natural ...
		}
		if exit := loop(db, login, strings.TrimSpace(cmd)); exit {
			return
		}
	}
}

func unauthorizedOperationsLoop(db *sql.DB, login string, cmd string) (exit bool) {
	switch cmd {
	case "1":
		ok, login, err := handleLogin(db)
		if err != nil {
			log.Printf("can't handle login: %v", err)
			return true
		}
		if !ok {
			fmt.Println("Неправильно введён логин или пароль. Попробуйте ещё раз. Для выхода введите буква q на логине и пароле! \n")
			unauthorizedOperationsLoop(db, "", "1")
			return true
		}
		operationsLoop(db, login, authorizedOperations, authorizedOperationsLoop)
		return false
	case "q":
		return true
	default:
		fmt.Printf("Вы выбрали неверную команду: %s\n", cmd)
	}

	return false
}

func handleLogin(db *sql.DB) (ok bool, loginClient string, err error) {
	fmt.Println("Введите ваш логин и пароль")
	var login string
	fmt.Print("Логин: ")
	_, err = fmt.Scan(&login)
	if err != nil {
		return false, login, err
	}
	var password string
	fmt.Print("Пароль: ")
	_, err = fmt.Scan(&password)
	if err != nil {
		return false, login, err
	}

	if login == "q" && password == "q" {
		return false, login, errors.New("want to exit")
	}

	ok, err = core.Login(login, password, "clients", db)
	if err != nil {
		return false, login, err
	}

	return ok, login, err
}

func authorizedOperationsLoop(db *sql.DB, login string, cmd string) (exit bool) {
	//ClearConsole()
	switch cmd {
	case "1":
		//-------1. Посмотреть список счетов
		var mainAcc int
		accounts := make([]core.Accounts, 0)
		err := showAccountClient(db, login, &mainAcc, &accounts)
		fmt.Printf("\n\n")
		if err != nil {
			fmt.Println("Сервер временно недоступен, повторите попытку позже")
			return false
		}
	case "2":
		//--------2. Оплатить услугу
		services := make([]core.Services, 0)

		err := core.ShowServices(db, &services)
		if err != nil {
			fmt.Println("Сервер временно недоступен, повторите попытку позже")
			return false
		}
		if len(services) == 0 {
			fmt.Println("Не одного услуга не найдена!")
			return false
		}
		fmt.Printf("Выберите соответствующий номер услуги из списка:\n")
		for i, val := range services {
			fmt.Printf("%d. Названия услуга: %s \n", i+1, val.Name)
		}
		fmt.Printf("%d. Назад \n", len(services)+1)
		var cmd, money int
		fmt.Print("Номер услуги: ")
		_, err = fmt.Scan(&cmd)
		if err != nil || cmd > len(services)+1 {
			fmt.Println("Неправелный ввод команды! Повторите занова!")
			authorizedOperationsLoop(db, login, "2")
			return false
		}
		if cmd == len(services)+1 {
			return false
		}
		var idPayment string
		fmt.Printf("Вводите %s: ", strings.ToLower(services[cmd-1].IdPayment))
		_, err1 := fmt.Scan(&idPayment)
		fmt.Print("Сумма: ")
		_, err = fmt.Scan(&money)
		if err != nil || money <= 0 || err1 != nil {
			fmt.Println("Неправелный ввод команды! Повторите занова!")
			authorizedOperationsLoop(db, login, "2")
			return false
		}

		fmt.Println("Выберите номер команды счета для оплаты суммы услуги:")
		account := make([]core.Accounts, 0)
		var mainAcc int
		err = showAccountClient(db, login, &mainAcc, &account)
		fmt.Printf(`%d) Назад к меню `, len(account)+1)
		fmt.Printf("\n\n")
		if err != nil {
			fmt.Println("Сервер временно недоступен, повторите попытку позже")
			return false
		}
		fmt.Print("Выберите соответсвующий номер карты из списка для получения денег: ")
		_, err = fmt.Scan(&cmd)
		if err != nil || cmd > len(account)+1 {
			fmt.Println("Неправелный ввод команды! Повторите занова!")
			authorizedOperationsLoop(db, login, "2")
			return false
		}
		if cmd == len(account)+1 {
			return false
		}
		if account[cmd-1].Money < money {
			fmt.Println("На выбранному номер счёта недостаточно средств для оплаты сумму услуги! Повторите заново!")
			authorizedOperationsLoop(db, login, "2")
			return false
		}
		err = core.AddMoneyAccountNumber(db, account[cmd-1].Number, -money)
		if err != nil {
			fmt.Println("Сервер временно недоступен, повторите попытку позже")
			return false
		}
		err = core.AddMoneyAccountNumber(db, services[cmd-1].AccountNumber, money)
		if err != nil {
			fmt.Println("Сервер временно недоступен, повторите попытку позже")
			return false
		}
		fmt.Print("Платеж успешно завершен\n\n")
	case "3":
		//--------3. Список банкоматов
		fmt.Println("        Список банкоматов:")
		terminal := make([]core.Terminals, 0)
		err := core.ShowTerminals(db, &terminal)
		if err != nil {
			return false
		}
		for i, val := range terminal {
			fmt.Printf(`%d)Адрес банкомата: "%s"`, i+1, val.Address)
			fmt.Println()
		}
		fmt.Printf("\n\n")
	case "4":
		//--------4. Перевести деньги другому клиенту
		transferMoney(db, login)
	case "5":
		//--------5. Перевести из одного счёт в другой счёт (свой)
		fmt.Println("Выберите из какого счёта на какой счёт хотите перевести денег:")
		var mainAcc int
		accounts := make([]core.Accounts, 0)
		err := showAccountClient(db, login, &mainAcc, &accounts)
		fmt.Printf(`%d) Назад к меню (Вводите на любой поля номер "%d" ) `, len(accounts)+1, len(accounts)+1)
		fmt.Printf("\n\n")
		if err != nil {
			fmt.Println("Сервер временно недоступен, повторите попытку позже")
			return false
		}
		var senAcc, recAcc, money int
		fmt.Print("Выберите соответсвующий номер карты из списка для перевода денег: ")
		_, err = fmt.Scan(&senAcc)
		if err != nil || senAcc > len(accounts)+1 {
			fmt.Println("Неправелный ввод команды! Повторите занова!")
			authorizedOperationsLoop(db, login, cmd)
			return false
		}
		if senAcc == len(accounts)+1 {
			return false
		}
		fmt.Print("Выберите соответсвующий номер карты из списка для получения денег: ")
		_, err = fmt.Scan(&recAcc)
		if err != nil || recAcc > len(accounts)+1 {
			fmt.Println("Неправелный ввод команды! Повторите занова!")
			authorizedOperationsLoop(db, login, cmd)
			return false
		}
		if recAcc == len(accounts)+1 {
			return false
		}
		if recAcc == senAcc {
			fmt.Println("Нельзя перевести деньги на одну и ту же карту")
			authorizedOperationsLoop(db, login, cmd)
			return false
		}
		fmt.Print("Сумма: ")
		_, err = fmt.Scan(&money)
		if err != nil {
			fmt.Println("Неправелный ввод команды! Повторите занова!")
			authorizedOperationsLoop(db, login, cmd)
			return false
		}
		if money > accounts[senAcc-1].Money {
			fmt.Println("Недостаточно средств на счёту от которого хотите отправить! Выберите другой счет!")
			authorizedOperationsLoop(db, login, cmd)
			return false
		}
		if money <= 0 {
			fmt.Println("Неправильная ввода сумма перевода! Повторите занова!")
			authorizedOperationsLoop(db, login, cmd)
			return false
		}
		err = core.AddMoneyAccountNumber(db, accounts[senAcc-1].Number, -money)
		if err != nil {
			fmt.Println("Сервер временно недоступен, повторите попытку позже")
			return false
		}
		err = core.AddMoneyAccountNumber(db, accounts[recAcc-1].Number, money)
		if err != nil {
			fmt.Println("Сервер временно недоступен, повторите попытку позже")
			return false
		}
		fmt.Print("Платеж успешно завершен\n\n")
	case "6":
		//--------6. Выбрать счёт для пролучения перевода (перевод через номер телефона)
		fmt.Println("Выбирите номер команды счёта для установка получения перевода:")
		var mainAcc, choose int
		accounts := make([]core.Accounts, 0)
		err := showAccountClient(db, login, &mainAcc, &accounts)
		fmt.Printf("\n\n")
		if err != nil {
			fmt.Println("Сервер временно недоступен, повторите попытку позже")
			operationsLoop(db, login, authorizedOperations, authorizedOperationsLoop)
			log.Fatal(err)
			return false
		}
		fmt.Print("Номер команды:")
		_, err = fmt.Scan(&choose)
		if err != nil || choose > len(accounts) {
			fmt.Print("Неправильная ввода команда, повторите заново!\n\n")
			authorizedOperationsLoop(db, login, "6")
			log.Print(err, "Choosen command value is out of range ")
			return false
		}
		err = core.ChangeMainAcc(db, accounts[choose-1].ClientId, accounts[choose-1].Number)
		if err != nil {
			log.Print(err)
			return false
		}
		fmt.Println("Основная счёт было успешно изменено.\n")
	case "q":
		return true
	default:
		fmt.Printf("Вы выбрали неверную команду: %s\n", cmd)
	}
	return false
}

func showAccountClient(db *sql.DB, login string, mainAcc *int, accounts *[]core.Accounts) error {

	id, err := core.IdClientByLogin(db, login)
	if err != nil {
		fmt.Println("Сервер временно недоступен, повторите попытку позже")
		log.Print(err)
		return err
	}

	err = core.ShowAccountById(db, accounts, mainAcc, id)

	if err != nil {
		fmt.Println("Сервер временно недоступен, повторите попытку позже")
		log.Print(err)
		return err
	}

	for i, val := range *accounts {
		if val.Number == *mainAcc {
			fmt.Printf(`%v) Имя:"%v" Номер-счёт:"%v" Баланс:"%v"" (Выбрано для получения перевод денег)`, i+1, val.Name, val.Number, val.Money)
		} else {
			fmt.Printf(`%v) Имя:"%v" Номер-счёт:"%v" Баланс:"%v""`, i+1, val.Name, val.Number, val.Money)
		}
		fmt.Println()
	}
	return nil
}

func transferMoney(db *sql.DB, login string) {
	for {
		var cmd string
		var phone, numberAcc int
		fmt.Print("Выбирите команду:\n 1) Перевод денег по номеру телефона пользователя \n 2) Перевод денег по определенному номеру счёта \n q) Назад\n")
		_, err := fmt.Scan(&cmd)
		if err != nil {
			fmt.Println("Ввод неправильно! Повторите заново!")
			continue
		}
		if cmd == "1" {
			fmt.Println("Введите номер телефон пользователя:  ")
			_, err = fmt.Scan(&phone)
			if err != nil {
				fmt.Println("Ввод неправильно! Повторите заново!")
				continue
			}
			numberAcc, err := core.NumberAccountByPhone(db, int(phone))
			if err != nil {
				if err.Error() == "Undefined phone" {
					fmt.Println("Номер телефона не найден! Повторите еще раз!")
				} else {
					fmt.Println("Сервер недоступенно! Повторите заново!")
				}
				continue
			}
			payAskMoney(db, login, numberAcc)
			break
			///-------------------------------------------------------------------------
		} else if cmd == "2" {
			fmt.Println("Введите номер счёт пользователя:  ")
			_, err = fmt.Scan(&numberAcc)
			if err != nil {
				fmt.Println("Ввод неправильно! Повторите заново!")
				continue
			}
			err = core.CheckAccount(db, numberAcc)
			if err != nil {
				if err.Error() == "Undefined number account" {
					fmt.Println("Номер счёт пользователя не найден! Повторите еще раз!")
				} else {
					fmt.Println("Сервер недоступенно! Повторите заново!")
				}
				continue
			}
			payAskMoney(db, login, numberAcc)
			break
			///-------------------------------------------------------------------------
		} else if cmd == "q" {
			return
		} else {
			fmt.Println("Ввод команды неправильно! Повторите заново!")
		}
	}
}

func payAskMoney(db *sql.DB, login string, numberAcc int) {
	idNumberAcc, err := core.IdClientByAccount(db, numberAcc)
	idLogin, err1 := core.IdClientByLogin(db, login)
	if err != nil || err1 != nil {
		fmt.Println("Сервер временно недоступен, повторите попытку позже")
		return
	}
	fmt.Println(idLogin, idNumberAcc)
	if idNumberAcc == idLogin {
		fmt.Println("Нельзя совершать перевод денег сомому себе. Для этого эсть пункт в гловному меню (Перевести из одного счёт в другой счёт (свой))")
		return
	}
	var money, cmd int
	fmt.Println("Сумма перевода: ")
	_, err = fmt.Scan(&money)
	if err != nil || money <= 0 {
		fmt.Println("Неправелный ввод команды! Повторите занова!")
		payAskMoney(db, login, numberAcc)
		return
	}
	accounts := make([]core.Accounts, 0)
	var mainAcc int
	fmt.Println("Выберите номер команды счета для оплаты суммы услуги:")
	err = showAccountClient(db, login, &mainAcc, &accounts)
	fmt.Printf(`%d) Назад к меню `, len(accounts)+1)
	fmt.Printf("\n\n")
	if err != nil {
		fmt.Println("Сервер временно недоступен, повторите попытку позже")
		payAskMoney(db, login, numberAcc)
		return
	}
	fmt.Print("Выберите соответсвующий номер карты из списка для получения денег: ")
	_, err = fmt.Scan(&cmd)
	if err != nil || cmd > len(accounts)+1 {
		fmt.Println("Неправелный ввод команды! Повторите занова!")
		payAskMoney(db, login, numberAcc)
		return
	}
	if cmd == len(accounts)+1 {
		return
	}
	if accounts[cmd-1].Money < money {
		fmt.Println("На выбранному номер счёта недостаточно средств для оплаты сумму услуги! Повторите заново!")
		payAskMoney(db, login, numberAcc)
		return
	}
	err = core.AddMoneyAccountNumber(db, accounts[cmd-1].Number, -money)
	if err != nil {
		fmt.Println("Сервер временно недоступен, повторите попытку позже")
		return
	}
	err = core.AddMoneyAccountNumber(db, numberAcc, money)
	if err != nil {
		fmt.Println("Сервер временно недоступен, повторите попытку позже")
		return
	}
	fmt.Print("Платеж успешно завершен\n\n")
}
