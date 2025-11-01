package oracle

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	contracts "github.com/torvald2/wells_oracle/Contracts"
)

const (
	// Адрес вашего контракта
	ContractAddress = "0x495c16046E077EA4E6A71463De905BD8379eA49A"
	// ID подписки Chainlink Functions
	SubscriptionId = 5832
	// Периодичность опроса
	PollInterval = 30 * time.Minute
)

// Статический список ID для отправки в запросах
var staticIDs = []string{
	"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
}

// Service инкапсулирует логику для взаимодействия с контрактом Oracle.
type Service struct {
	client         *ethclient.Client
	contract       *contracts.Oracle
	contractFilter *contracts.OracleFilterer
	auth           *bind.TransactOpts
	// Потокобезопасное хранилище для результатов
	valuationData   map[string]string
	valuationDataMu sync.RWMutex
}

// NewService создает новый экземпляр сервиса для контракта Oracle.
// ethNodeURL: URL адрес узла Ethereum (например, "http://localhost:8545").
// privateKeyHex: Приватный ключ в виде шестнадцатеричной строки для подписи транзакций.
func NewService(ethNodeURL, privateKeyHex string) (*Service, error) {
	client, err := ethclient.Dial(ethNodeURL)
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к узлу Ethereum: %w", err)
	}

	contractAddress := common.HexToAddress(ContractAddress)
	contract, err := contracts.NewOracle(contractAddress, client)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать экземпляр контракта Oracle: %w", err)
	}

	contractFilter, err := contracts.NewOracleFilterer(contractAddress, client)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать фильтр событий для контракта: %w", err)
	}

	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("не удалось обработать приватный ключ: %w", err)
	}

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("не удалось получить chain ID: %w", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать авторизованный транзактор: %w", err)
	}

	return &Service{
		client:         client,
		contract:       contract,
		contractFilter: contractFilter,
		auth:           auth,
		valuationData:  make(map[string]string),
	}, nil
}

// StartPolling запускает бесконечный цикл для периодической отправки запросов.
func (s *Service) StartPolling(ctx context.Context) {
	log.Println("Запуск периодического опроса...")
	ticker := time.NewTicker(PollInterval)
	defer ticker.Stop()

	// Выполняем первый раз сразу, не дожидаясь тикера

	s.processAllIDs(ctx)

	for {
		select {
		case <-ticker.C:
			s.processAllIDs(ctx)
		case <-ctx.Done():
			log.Println("Остановка опроса.")
			return
		}
	}
}

// processAllIDs отправляет запросы для всех ID из статического списка.
func (s *Service) processAllIDs(ctx context.Context) {
	log.Println("Отправка запросов для всех ID...")
	for _, id := range staticIDs {
		log.Printf("Отправка запроса для ID: %s\n", id)
		args := []string{id}
		tx, err := s.SendRequest(SubscriptionId, args)
		if err != nil {
			log.Printf("Ошибка отправки запроса для ID %s: %v\n", id, err)
			continue
		}

		log.Printf("Транзакция отправлена: %s\n", tx.Hash().Hex())

		// Ждем подтверждения транзакции
		receipt, err := bind.WaitMined(ctx, s.client, tx)
		if err != nil {
			log.Printf("Ошибка ожидания майнинга транзакции для ID %s: %v\n", id, err)
			continue
		}

		if receipt.Status != types.ReceiptStatusSuccessful {
			log.Printf("Транзакция не удалась для ID %s\n", id)
			continue
		}

		log.Printf("Транзакция успешна для ID %s\n", id)

		// Получаем данные из контракта
		data, err := s.GetValuationData(id)
		if err != nil {
			log.Printf("Ошибка получения данных для ID %s: %v\n", id, err)
			continue
		}

		// Сохраняем данные
		s.valuationDataMu.Lock()
		s.valuationData[id] = string(data)
		s.valuationDataMu.Unlock()

		log.Printf("Данные сохранены для ID %s: %s\n", id, string(data))
	}
}

// SendRequest отправляет запрос в смарт-контракт.
func (s *Service) SendRequest(subscriptionId uint64, args []string) (*types.Transaction, error) {
	tx, err := s.contract.SendRequest(s.auth, subscriptionId, args)
	if err != nil {
		return nil, fmt.Errorf("не удалось отправить транзакцию SendRequest: %w", err)
	}
	return tx, nil
}

// GetValuationData получает данные оценки по ID.
func (s *Service) GetValuationData(id string) ([]byte, error) {
	// Здесь нужно реализовать логику получения данных из контракта
	// Предполагается, что есть функция в контракте, которая возвращает данные по ID
	// Например, SValuationData или аналогичная
	// Для примера используем SLastResponse как заглушку
	response, err := s.contract.SLastResponse(nil)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить SLastResponse: %w", err)
	}
	return response, nil
}

// GetLastResponse получает последний ответ от контракта.
func (s *Service) GetLastResponse() ([]byte, error) {
	// Для вызовов view-функций можно использовать nil вместо opts, чтобы использовать последние данные блокчейна.
	response, err := s.contract.SLastResponse(nil)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить SLastResponse: %w", err)
	}
	return response, nil
}

// GetLastError получает последнюю ошибку от контракта.
func (s *Service) GetLastError() ([]byte, error) {
	er, err := s.contract.SLastError(nil)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить SLastError: %w", err)
	}
	return er, nil
}

// GetOwner получает владельца контракта.
func (s *Service) GetOwner() (common.Address, error) {
	owner, err := s.contract.Owner(nil)
	if err != nil {
		return common.Address{}, fmt.Errorf("не удалось получить Owner: %w", err)
	}
	return owner, nil
}

func (s *Service) GetValuationDataSaved(id string) string {
	s.valuationDataMu.Lock()
	defer s.valuationDataMu.Unlock()

	data := s.valuationData[id]

	return data
}
