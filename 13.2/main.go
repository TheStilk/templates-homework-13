package main

import (
	"errors"
	"fmt"
)

// States
type State interface {
	SelectTicket(m *TicketMachine, ticketType string) error
	InsertMoney(m *TicketMachine, amount float64) error
	Cancel(m *TicketMachine) error
	DispenseTicket(m *TicketMachine) error
	Name() string
}

type IdleState struct{}

func (s *IdleState) SelectTicket(m *TicketMachine, ticketType string) error {
	if !m.HasTicket(ticketType) {
		return errors.New("ticket unavailable")
	}
	m.CurrentTicket = ticketType
	m.CurrentPrice = m.GetTicketPrice(ticketType)
	m.SetState(&WaitingForMoneyState{})
	fmt.Printf("Ticket selected: %s (%.2f KZT)\n", ticketType, m.CurrentPrice)
	return nil
}

func (s *IdleState) InsertMoney(m *TicketMachine, amount float64) error {
	return errors.New("please select a ticket first")
}
func (s *IdleState) Cancel(m *TicketMachine) error {
	return errors.New("no active transaction")
}
func (s *IdleState) DispenseTicket(m *TicketMachine) error {
	return errors.New("no paid ticket")
}
func (s *IdleState) Name() string { return "Idle" }

type WaitingForMoneyState struct{}

func (s *WaitingForMoneyState) SelectTicket(m *TicketMachine, ticketType string) error {
	return errors.New("ticket already selected")
}

func (s *WaitingForMoneyState) InsertMoney(m *TicketMachine, amount float64) error {
	m.InsertedMoney += amount
	fmt.Printf("Inserted: %.2f KZT (Total: %.2f)\n", amount, m.InsertedMoney)
	if m.InsertedMoney >= m.CurrentPrice {
		m.SetState(&MoneyReceivedState{})
		fmt.Println("Sufficient funds. Ready to dispense ticket.")
	}
	return nil
}

func (s *WaitingForMoneyState) Cancel(m *TicketMachine) error {
	m.SetState(&TransactionCanceledState{})
	return nil
}

func (s *WaitingForMoneyState) DispenseTicket(m *TicketMachine) error {
	return errors.New("insufficient funds")
}
func (s *WaitingForMoneyState) Name() string { return "WaitingForMoney" }

type MoneyReceivedState struct{}

func (s *MoneyReceivedState) SelectTicket(m *TicketMachine, ticketType string) error {
	return errors.New("ticket already selected")
}

func (s *MoneyReceivedState) InsertMoney(m *TicketMachine, amount float64) error {
	m.InsertedMoney += amount
	fmt.Printf("Additional funds inserted: %.2f KZT\n", amount)
	return nil
}

func (s *MoneyReceivedState) Cancel(m *TicketMachine) error {
	m.SetState(&TransactionCanceledState{})
	return nil
}

func (s *MoneyReceivedState) DispenseTicket(m *TicketMachine) error {
	m.SetState(&TicketDispensedState{})
	m.Inventory[m.CurrentTicket]--
	m.InsertedMoney = 0
	m.CurrentTicket = ""
	fmt.Println("Ticket dispensed!")
	return nil
}
func (s *MoneyReceivedState) Name() string { return "MoneyReceived" }

type TicketDispensedState struct{}

func (s *TicketDispensedState) handle() {}

func (s *TicketDispensedState) SelectTicket(m *TicketMachine, ticketType string) error {
	return errors.New("please take your ticket and start over")
}
func (s *TicketDispensedState) InsertMoney(m *TicketMachine, amount float64) error {
	return errors.New("please take your ticket")
}
func (s *TicketDispensedState) Cancel(m *TicketMachine) error {
	return errors.New("transaction complete")
}
func (s *TicketDispensedState) DispenseTicket(m *TicketMachine) error {
	return errors.New("ticket already dispensed")
}
func (s *TicketDispensedState) Name() string { return "TicketDispensed" }

type TransactionCanceledState struct{}

func (s *TransactionCanceledState) handle() {}

func (s *TransactionCanceledState) SelectTicket(m *TicketMachine, ticketType string) error {
	return errors.New("transaction canceled. Please start over")
}
func (s *TransactionCanceledState) InsertMoney(m *TicketMachine, amount float64) error {
	return errors.New("transaction canceled")
}
func (s *TransactionCanceledState) Cancel(m *TicketMachine) error {
	return errors.New("already canceled")
}
func (s *TransactionCanceledState) DispenseTicket(m *TicketMachine) error {
	return errors.New("no ticket")
}
func (s *TransactionCanceledState) Name() string { return "TransactionCanceled" }

// Machine

type TicketMachine struct {
	State         State
	CurrentTicket string
	CurrentPrice  float64
	InsertedMoney float64
	Inventory     map[string]int
	TicketPrices  map[string]float64
}

func NewTicketMachine() *TicketMachine {
	return &TicketMachine{
		State:        &IdleState{},
		Inventory:    map[string]int{"metro": 10, "bus": 15, "train": 5},
		TicketPrices: map[string]float64{"metro": 300.0, "bus": 250.0, "train": 1000.0},
	}
}

func (m *TicketMachine) SetState(s State) {
	m.State = s
}

func (m *TicketMachine) GetCurrentState() string {
	return m.State.Name()
}

func (m *TicketMachine) GetTicketPrice(ticketType string) float64 {
	return m.TicketPrices[ticketType]
}

func (m *TicketMachine) HasTicket(ticketType string) bool {
	return m.Inventory[ticketType] > 0
}

func (m *TicketMachine) SelectTicket(ticketType string) error {
	return m.State.SelectTicket(m, ticketType)
}

func (m *TicketMachine) InsertMoney(amount float64) error {
	return m.State.InsertMoney(m, amount)
}

func (m *TicketMachine) Cancel() error {
	return m.State.Cancel(m)
}

func (m *TicketMachine) DispenseTicket() error {
	return m.State.DispenseTicket(m)
}

func main() {
	machine := NewTicketMachine()

	fmt.Println("--- Successful Purchase ---")
	machine.SelectTicket("metro")
	machine.InsertMoney(300.0)
	machine.DispenseTicket()

	fmt.Println("\n--- Cancellation Before Payment ---")
	machine = NewTicketMachine()
	machine.SelectTicket("bus")
	machine.Cancel()
	fmt.Printf("State: %s\n", machine.GetCurrentState())

	fmt.Println("\n--- Cancellation After Payment ---")
	machine = NewTicketMachine()
	machine.SelectTicket("train")
	machine.InsertMoney(1000.0)
	machine.Cancel()
	fmt.Printf("State: %s\n", machine.GetCurrentState())
}
