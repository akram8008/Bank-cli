package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/akram8008/Bank-core/core"
	"io"
	"io/ioutil"
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
	operationsLoop(db, unauthorizedOperations, unauthorizedOperationsLoop)
	log.Print("finish operations loop")
	log.Print("finish application")
}

func operationsLoop(db *sql.DB, commands string, loop func(db *sql.DB, cmd string) bool) {
	for {
		fmt.Println(commands)
		var cmd string
		_, err := fmt.Scan(&cmd)
		if err != nil {
			log.Fatalf("Can't read input: %v", err) // %v - natural ...
		}
		if exit := loop(db, strings.TrimSpace(cmd)); exit {
			return
		}
	}
}

func unauthorizedOperationsLoop(db *sql.DB, cmd string) (exit bool) {
	switch cmd {
	case "1":
		ok, err := handleLogin(db)
		if err != nil {
			log.Printf("can't handle login: %v", err)
			return true
		}
		if !ok {
			fmt.Println("Неправильно введён логин или пароль. Попробуйте ещё раз. Для выхода введите буква q на логине и пароле! \n")
			unauthorizedOperationsLoop(db, "1")
			return true
		}
		operationsLoop(db, authorizedOperations, authorizedOperationsLoop)
		return false
	case "q":
		return true
	default:
		fmt.Printf("Вы выбрали неверную команду: %s\n", cmd)
	}

	return false
}

func handleLogin(db *sql.DB) (ok bool, err error) {
	fmt.Println("Введите ваш логин и пароль")
	var login string
	fmt.Print("Логин: ")
	_, err = fmt.Scan(&login)
	if err != nil {
		return false, err
	}
	var password string
	fmt.Print("Пароль: ")
	_, err = fmt.Scan(&password)
	if err != nil {
		return false, err
	}

	if login == "q" && password == "q" {
		return false, errors.New("want to exit")
	}

	ok, err = core.Login(login, password, "managers", db)
	if err != nil {
		return false, err
	}

	return ok, err
}

func authorizedOperationsLoop(db *sql.DB, cmd string) (exit bool) {
	switch cmd {
	case "1":
		//-------1. Добавить нового пользователя
		newClient, err := scanNewClient()
		if err != nil {
			fmt.Printf("Поля были заполнены неправильно, попробуйте еще раз.")
			log.Printf("During registration a new client was detected: %v", err)
			authorizedOperationsLoop(db, "1")
			return false
		}
		numberAccount, err := core.AddNewClient(db, newClient)
		if err != nil {
			if err.Error() == "UNIQUE constraint failed: clients.login" {
				fmt.Printf("Позватель с таким логином уже существует. Попробуйте заново. \n ")
			} else if err.Error() == "UNIQUE constraint failed: clients.serialPass" {
				fmt.Printf("Позватель с таким серия паспорта уже существует. Попробуйте заново. \n ")
			} else if err.Error() == "UNIQUE constraint failed: clients.phone" {
				fmt.Printf("Позватель с таким номер телфона уже существует. Попробуйте заново. \n ")
			} else {
				fmt.Printf("Сервер недоступен. Попробуйте заново. \n")
			}
			log.Printf("Has finished with error: %v", err)
			authorizedOperationsLoop(db, "1")
			return false
		}
		fmt.Printf("Новый пользователь успешно добавлен. Номер первый счет: %d \n", numberAccount)
		log.Printf("A new client has been added and number account is  %d \n", numberAccount)
	case "2":
		//--------2. Добавить дополнительное счёт пользователю
		var choose, login, phone, name string
		fmt.Printf("Добавление новый счёт, ")
		fmt.Println(selectLoginPhone)
		log.Printf("Adding a new account to user")
		_, err := fmt.Scan(&choose)
		if err != nil {
			fmt.Printf("Поля были заполнены неправильно, попробуйте еще раз.\n")
			log.Printf("During registration a new account was detected: %v", err)
			authorizedOperationsLoop(db, "2")
			return false
		}
		switch choose {
		case "1":
			fmt.Println("Логин:")
			_, err := fmt.Scan(&login)
			if err != nil {
				fmt.Printf("Поля были заполнены неправильно, попробуйте еще раз.\n")
				log.Printf("During registration a new account was detected: %v", err)
				authorizedOperationsLoop(db, "2")
				return false
			}
			fmt.Println("Имя счёта:")
			_, err = fmt.Scan(&name)
			if err != nil {
				fmt.Printf("Поля были заполнены неправильно, попробуйте еще раз.\n")
				log.Printf("During registration a new account was detected: %v", err)
				authorizedOperationsLoop(db, "2")
				return false
			}
			numberAcc, err := core.AddAccountByLogin(db, login, name)
			if err != nil {
				fmt.Printf("Заданый логин ненайден, попробуйте еще раз.\n")
				log.Printf("Couldn't find login in DataBase: %v", err)
				authorizedOperationsLoop(db, "2")
				return false
			}
			fmt.Printf("%v с номерам счета:%v успешно созданно\n", name, numberAcc)
			return false
		case "2":
			fmt.Println("Номер телфон:")
			_, err := fmt.Scan(&phone)
			if err != nil {
				fmt.Printf("Поля были заполнены неправильно, попробуйте еще раз.\n")
				log.Printf("During registration a new account was detected: %v", err)
				authorizedOperationsLoop(db, "2")
				return false
			}
			fmt.Println("Имя счёта:")
			_, err = fmt.Scan(&name)
			if err != nil {
				fmt.Printf("Поля были заполнены неправильно, попробуйте еще раз.\n")
				log.Printf("During registration a new account was detected: %v", err)
				authorizedOperationsLoop(db, "2")
				return false
			}
			numberAcc, err := core.AddAccountByPhone(db, phone, name)
			if err != nil {
				fmt.Printf("Заданый номер телфон ненайден, попробуйте еще раз.\n")
				log.Printf("Couldn't find login in DataBase: %v", err)
				authorizedOperationsLoop(db, "2")
				return false
			}
			fmt.Printf("%v с номерам счета:%v успешно созданно\n\n", name, numberAcc)
			return false
		case "q":
			//operationsLoop(db, authorizedOperations, authorizedOperationsLoop)
			return false
		default:
			fmt.Printf("Неправильно выбранная команда \n ")
			authorizedOperationsLoop(db, "2")
			return false
		}
	case "3":
		//--------3. Пополнить баланс пользователя
		payAccount(db)
	case "4":
		//--------4. Добавить услуги (название)
		var name string
		var number string
		fmt.Printf("			Добавление нового услуга\n")
		fmt.Printf("Введите имя нового услуга:")
		name, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Поля были заполнены неправильно, попробуйте еще раз.\n")
			log.Printf("During registration a new account was detected: %v", err)
			authorizedOperationsLoop(db, "7")
			return false
		}
		fmt.Printf("Введите идентификационный названия продукта:")
		number, err = reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Поля были заполнены неправильно, попробуйте еще раз.\n")
			log.Printf("During registration a new account was detected: %v", err)
			authorizedOperationsLoop(db, "7")
			return false
		}

		err = core.AddServices(db, name, number)
		if err != nil {
			if err.Error() == "Service already exits" {
				fmt.Printf("Услуга уже зарегистрирован")
			} else {
				fmt.Printf("Сервер недоступен. Попробуйте заново. \n")
			}
			return false
		}
		fmt.Printf("Услуга с успешно зарегистрирован\n \n")
		return false
	case "5":
		//--------5. Экспорт (форматы json и xml)
		for {
			var num int
			fmt.Println(exportList)
			_, err := fmt.Scanln(&num)
			if err != nil {
				fmt.Println("Ввод неправелный! Вводите заново!")
				continue
			}
			if num == 1 {
				err = exportClients(db)
				if err != nil {
					fmt.Println("Сервер временно недоступен!")
					return false
				}
				break
			} else if num == 2 {
				err = exportAccounts(db)
				if err != nil {
					fmt.Println("Сервер временно недоступен!")
					return false
				}
				break
			} else if num == 3 {
				err = exporTerminals(db)
				if err != nil {
					fmt.Println("Сервер временно недоступен!")
					return false
				}
				break
			} else if num == 4 {
				break
			} else {
				fmt.Println("Ввод неправелный! Вводите заново!")
			}
		}
	case "6":
		//--------6. Импорт  (форматы json и xml)
		for {
			var num int
			fmt.Println(importList)
			_, err := fmt.Scanln(&num)
			if err != nil {
				fmt.Println("Ввод неправелный! Вводите заново!")
				continue
			}
			if num == 1 {
				err = importClients(db, "json")
				if err != nil {
					fmt.Println("Сервер временно недоступен!")
					return false
				}
				break
			} else if num == 2 {
				err = importClients(db, "xml")
				if err != nil {
					fmt.Println("Сервер временно недоступен!")
					return false
				}
				break
			} else if num == 3 {
				err = importAccounts(db, "json")
				if err != nil {
					fmt.Println("Сервер временно недоступен!")
					return false
				}
				break
			} else if num == 4 {
				err = importAccounts(db, "xml")
				if err != nil {
					fmt.Println("Сервер временно недоступен!")
					return false
				}
				break
			} else if num == 5 {
				err = importTerminals(db, "json")
				if err != nil {
					fmt.Println("Сервер временно недоступен!")
					return false
				}
				break
			} else if num == 6 {
				err = importTerminals(db, "xml")
				if err != nil {
					fmt.Println("Сервер временно недоступен!")
					return false
				}
				break
			} else if num == 7 {
				break
			} else {
				fmt.Println("Ввод неправелный! Вводите заново!")
			}
		}
	case "7":
		//--------7. Добавить банкомат
		var number, address string
		fmt.Printf("Добавление нового банкомата:\n")

		fmt.Printf("Введите идентификационный номер банкомата:")
		_, err := fmt.Scan(&number)
		if err != nil {
			fmt.Printf("Поля были заполнены неправильно, попробуйте еще раз.\n")
			log.Printf("During registration a new account was detected: %v", err)
			authorizedOperationsLoop(db, "7")
			return false
		}
		fmt.Printf("Введите адрес банкомата:")
		_, err = fmt.Scan(&address)
		if err != nil {
			fmt.Printf("Поля были заполнены неправильно, попробуйте еще раз.\n")
			log.Printf("During registration a new account was detected: %v", err)
			authorizedOperationsLoop(db, "7")
			return false
		}

		err = core.AddTerminals(db, number, address)
		if err != nil {
			if err.Error() == "Terminal already exits" {
				fmt.Printf("Банкомат с таким номер уже зарегистрирован")
			} else {
				fmt.Printf("Сервер недоступен. Попробуйте заново. \n")
			}
			authorizedOperationsLoop(db, "7")
			return false
		}
		fmt.Printf("Банкомат с успешно зарегистрирован\n \n")
		return false
	case "q":
		return true
	default:
		fmt.Printf("Вы выбрали неверную команду: %s\n", cmd)
	}
	return false
}

func payNumberAccount(db *sql.DB, numberAcc int) error {
	err := core.CheckAccount(db, numberAcc)
	if err != nil {
		if err.Error() == "Undefined number account" {
			fmt.Println("Номер счёт пользователя не найден! Повторите еще раз!")
		}
		return err
	}

	fmt.Println("Введите сумму: ")
	var amount int
	_, err = fmt.Scan(&amount)
	if err != nil {
		fmt.Println("Ввод неправильно! Повторите заново!")
		return err
	}
	err = core.AddMoneyAccountNumber(db, numberAcc, amount)
	if err != nil {
		fmt.Println("Сервер недоступенно! Повторите заново!")
		return err
	}
	return nil
}

func payAccount(db *sql.DB) {
	for {
		var cmd string
		var phone, numberAcc int
		fmt.Print("Выбирите команду:\n 1) По номеру телефона пользователя \n 2) По определенному номеру счёта \n q) Назад\n")
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
			err = payNumberAccount(db, numberAcc)
			if err != nil {
				continue
			}
			fmt.Print("Платеж успешно завершен\n\n")
			break
			///-------------------------------------------------------------------------
		} else if cmd == "2" {
			fmt.Println("Введите номер счёт пользователя:  ")
			_, err = fmt.Scan(&numberAcc)
			if err != nil {
				fmt.Println("Ввод неправильно! Повторите заново!")
				continue
			}
			err = payNumberAccount(db, numberAcc)
			if err != nil {
				continue
			}
			fmt.Print("Платеж успешно завершен\n\n")
			break
			///-------------------------------------------------------------------------
		} else if cmd == "q" {
			return
		} else {
			fmt.Println("Ввод команды неправильно! Повторите заново!")
		}
	}
}

func scanNewClient() (core.Client, error) {
	newClient := core.Client{}
	fmt.Println("Заполните информацию о новом пользователе:")
	fmt.Print("Имя:")
	_, err := fmt.Scan(&newClient.Name)
	if err != nil {
		return core.Client{}, err
	}
	fmt.Print("Фамилия:")
	_, err = fmt.Scan(&newClient.Surname)
	if err != nil {
		return core.Client{}, err
	}
	fmt.Print("Серия паспорта:")
	_, err = fmt.Scan(&newClient.SerialPass)
	if err != nil {
		return core.Client{}, err
	}
	fmt.Print("Логин:")
	_, err = fmt.Scan(&newClient.Login)
	if err != nil {
		return core.Client{}, err
	}
	fmt.Print("Пароль:")
	_, err = fmt.Scan(&newClient.Password)
	if err != nil {
		return core.Client{}, err
	}
	fmt.Print("Номер телефон:")
	_, err = fmt.Scan(&newClient.Phone)
	if err != nil {
		return core.Client{}, err
	}

	return newClient, nil
}

//---------------Exports-------------------
func exporTerminals(db *sql.DB) error {
	terminals := make([]core.Terminals, 0)
	err := core.ShowTerminals(db, &terminals)
	if err != nil {
		return err
	}
	jsonT, err := json.Marshal(terminals)
	if err != nil {
		return err
	}
	xmlT, err := xml.Marshal(terminals)
	if err != nil {
		return err
	}
	fileJ, err := os.OpenFile("terminals.json", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		return errors.New("Can not open for export")
	}
	fileX, err := os.OpenFile("terminals.xml", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		return errors.New("Can not open for export")
	}
	defer func() {
		fileJ.Close()
		fileX.Close()
	}()
	if err != nil {
		return errors.New("Can not open for export")
	}
	_, err = fileJ.WriteString(string(jsonT))
	if err != nil {
		return errors.New("Can not open for export")
	}
	_, err = fileX.WriteString(string(xmlT))
	if err != nil {
		return errors.New("Can not open the file for export")
	}
	fmt.Printf("Экспортирования успешно завершено!\n\n")
	return nil
}

func exportAccounts(db *sql.DB) error {
	accounts := make([]core.Accounts, 0)
	err := core.ShowAccounts(db, &accounts)
	if err != nil {
		return err
	}
	jsonT, err := json.Marshal(accounts)
	if err != nil {
		return err
	}
	xmlT, err := xml.Marshal(accounts)
	if err != nil {
		return err
	}
	fileJ, err := os.OpenFile("accounts.json", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		return errors.New("Can not open for export")
	}
	fileX, err := os.OpenFile("accounts.xml", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		return errors.New("Can not open for export")
	}
	defer func() {
		fileJ.Close()
		fileX.Close()
	}()
	if err != nil {
		return errors.New("Can not open for export")
	}
	_, err = fileJ.WriteString(string(jsonT))
	if err != nil {
		return errors.New("Can not open for export")
	}
	_, err = fileX.WriteString(string(xmlT))
	if err != nil {
		return errors.New("Can not open the file for export")
	}
	fmt.Printf("Экспортирования успешно завершено!\n\n")
	return nil
}

func exportClients(db *sql.DB) error {
	clients := make([]core.Client, 0)
	err := core.ShowClients(db, &clients)
	if err != nil {
		return err
	}
	jsonT, err := json.Marshal(clients)
	if err != nil {
		return err
	}
	xmlT, err := xml.Marshal(clients)
	if err != nil {
		return err
	}
	fileJ, err := os.OpenFile("clients.json", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		return errors.New("Can not open for export")
	}
	fileX, err := os.OpenFile("clients.xml", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		return errors.New("Can not open for export")
	}
	defer func() {
		fileJ.Close()
		fileX.Close()
	}()
	if err != nil {
		return errors.New("Can not open for export")
	}
	_, err = fileJ.WriteString(string(jsonT))
	if err != nil {
		return errors.New("Can not open for export")
	}
	_, err = fileX.WriteString(string(xmlT))
	if err != nil {
		return errors.New("Can not open the file for export")
	}
	fmt.Printf("Экспортирования успешно завершено!\n\n")
	return nil
}

//---------------Import---------------------
func importTerminals(db *sql.DB, tp string) error {
	var str string
	if tp == "json" {
		str = `C:\Go\Bank\Bank-cli\terminals.json`
	} else {
		str = `C:\Go\Bank\Bank-cli\terminals.xml`
	}

	bytes, err := ioutil.ReadFile(str)
	if err != nil {
		log.Printf("error while reading file: %v", err)
		return err
	}
	var terminals []core.Terminals
	if tp == "json" {
		err = json.Unmarshal(bytes, &terminals)
	} else {
		err = xml.Unmarshal(bytes, &terminals)
	}
	if err != nil {
		return nil
	}

	err = core.UpdateTerminals(db, &terminals)
	if err != nil {
		return errors.New("Can connect with the server!")
	}
	fmt.Printf("Импортирования успешно завершено!\n\n")
	return nil
}

func importAccounts(db *sql.DB, tp string) error {

	var str string
	if tp == "json" {
		str = `C:\Go\Bank\Bank-cli\accounts.json`
	} else {
		str = `C:\Go\Bank\Bank-cli\accounts.xml`
	}

	bytes, err := ioutil.ReadFile(str)
	if err != nil {
		log.Printf("error while reading file: %v", err)
		return err
	}
	var accounts []core.Accounts
	if tp == "json" {
		err = json.Unmarshal(bytes, &accounts)
	} else {
		err = xml.Unmarshal(bytes, &accounts)
	}

	if err != nil {
		return nil
	}

	err = core.UpdateAccounts(db, &accounts)
	if err != io.EOF {
		return err
	}

	if err != nil {
		return errors.New("Can connect with the server!")
	}
	fmt.Printf("Импортирования успешно завершено!\n\n")
	return nil
}

func importClients(db *sql.DB, tp string) error {
	var str string
	if tp == "json" {
		str = `C:\Go\Bank\Bank-cli\clients.json`
	} else {
		str = `C:\Go\Bank\Bank-cli\clients.xml`
	}

	bytes, err := ioutil.ReadFile(str)
	if err != nil {
		log.Printf("error while reading file: %v", err)
		return err
	}
	var clients []core.Client

	if tp == "json" {
		err = json.Unmarshal(bytes, &clients)
	} else {
		err = xml.Unmarshal(bytes, &clients)
	}
	fmt.Printf("%v\n", string(bytes))
	if err != nil {
		return nil
	}

	err = core.UpdateClients(db, &clients)
	if err != nil {
		return errors.New("Can connect with the server!")
	}
	fmt.Printf("Импортирования успешно завершено!\n\n")
	return nil
}

//------------------------------------------

var reader = bufio.NewReader(os.Stdin)
