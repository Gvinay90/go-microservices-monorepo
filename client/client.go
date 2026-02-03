package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	expensepb "github.com/vinay/splitwise-grpc/proto/expense"
	userpb "github.com/vinay/splitwise-grpc/proto/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	defaultUserServiceAddr    = "localhost:50051"
	defaultExpenseServiceAddr = "localhost:50052"
)

func main() {
	userServiceAddr := os.Getenv("USER_SERVICE_ADDR")
	if userServiceAddr == "" {
		userServiceAddr = defaultUserServiceAddr
	}

	expenseServiceAddr := os.Getenv("EXPENSE_SERVICE_ADDR")
	if expenseServiceAddr == "" {
		expenseServiceAddr = defaultExpenseServiceAddr
	}

	fmt.Printf("Connecting to User Service at %s...\n", userServiceAddr)
	// Connect to User Service
	userConn, err := grpc.Dial(userServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to user service: %v", err)
	}
	defer userConn.Close()
	userClient := userpb.NewUserServiceClient(userConn)

	fmt.Printf("Connecting to Expense Service at %s...\n", expenseServiceAddr)
	// Connect to Expense Service
	expenseConn, err := grpc.Dial(expenseServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to expense service: %v", err)
	}
	defer expenseConn.Close()
	expenseClient := expensepb.NewExpenseServiceClient(expenseConn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("=== Splitwise-like Expense Sharing Demo ===")

	// 1. Register users
	fmt.Println("1. Registering users...")
	alice, err := userClient.RegisterUser(ctx, &userpb.RegisterUserRequest{
		Name:  "Alice",
		Email: "alice@example.com",
	})
	if err != nil {
		log.Fatalf("failed to register Alice: %v", err)
	}
	fmt.Printf("   ✓ Registered: %s (ID: %s)\n", alice.User.Name, alice.User.Id)

	bob, err := userClient.RegisterUser(ctx, &userpb.RegisterUserRequest{
		Name:  "Bob",
		Email: "bob@example.com",
	})
	if err != nil {
		log.Fatalf("failed to register Bob: %v", err)
	}
	fmt.Printf("   ✓ Registered: %s (ID: %s)\n", bob.User.Name, bob.User.Id)

	charlie, err := userClient.RegisterUser(ctx, &userpb.RegisterUserRequest{
		Name:  "Charlie",
		Email: "charlie@example.com",
	})
	if err != nil {
		log.Fatalf("failed to register Charlie: %v", err)
	}
	fmt.Printf("   ✓ Registered: %s (ID: %s)\n\n", charlie.User.Name, charlie.User.Id)

	// 2. Add friendships
	fmt.Println("2. Adding friendships...")
	_, err = userClient.AddFriend(ctx, &userpb.AddFriendRequest{
		UserId:   alice.User.Id,
		FriendId: bob.User.Id,
	})
	if err != nil {
		log.Fatalf("failed to add friend: %v", err)
	}
	fmt.Printf("   ✓ Alice and Bob are now friends\n")

	_, err = userClient.AddFriend(ctx, &userpb.AddFriendRequest{
		UserId:   alice.User.Id,
		FriendId: charlie.User.Id,
	})
	if err != nil {
		log.Fatalf("failed to add friend: %v", err)
	}
	fmt.Printf("   ✓ Alice and Charlie are now friends\n")

	_, err = userClient.AddFriend(ctx, &userpb.AddFriendRequest{
		UserId:   bob.User.Id,
		FriendId: charlie.User.Id,
	})
	if err != nil {
		log.Fatalf("failed to add friend: %v", err)
	}
	fmt.Printf("   ✓ Bob and Charlie are now friends\n\n")

	// 3. Create expenses
	fmt.Println("3. Creating expenses...")

	// Alice pays for dinner, split equally among all three
	expense1, err := expenseClient.CreateExpense(ctx, &expensepb.CreateExpenseRequest{
		Description: "Dinner at Italian Restaurant",
		TotalAmount: 90.00,
		PaidBy:      alice.User.Id,
		SplitType:   expensepb.SplitType_EQUAL,
		Splits: []*expensepb.Split{
			{UserId: alice.User.Id, Amount: 30.00},
			{UserId: bob.User.Id, Amount: 30.00},
			{UserId: charlie.User.Id, Amount: 30.00},
		},
	})
	if err != nil {
		log.Fatalf("failed to create expense: %v", err)
	}
	fmt.Printf("   ✓ Expense created: %s ($%.2f)\n", expense1.Expense.Description, expense1.Expense.TotalAmount)
	fmt.Printf("     Paid by Alice, split equally\n")

	// Bob pays for movie tickets, split equally between Bob and Charlie
	expense2, err := expenseClient.CreateExpense(ctx, &expensepb.CreateExpenseRequest{
		Description: "Movie Tickets",
		TotalAmount: 40.00,
		PaidBy:      bob.User.Id,
		SplitType:   expensepb.SplitType_EQUAL,
		Splits: []*expensepb.Split{
			{UserId: bob.User.Id, Amount: 20.00},
			{UserId: charlie.User.Id, Amount: 20.00},
		},
	})
	if err != nil {
		log.Fatalf("failed to create expense: %v", err)
	}
	fmt.Printf("   ✓ Expense created: %s ($%.2f)\n", expense2.Expense.Description, expense2.Expense.TotalAmount)
	fmt.Printf("     Paid by Bob, split with Charlie\n")

	// Charlie pays for groceries, custom split
	expense3, err := expenseClient.CreateExpense(ctx, &expensepb.CreateExpenseRequest{
		Description: "Groceries",
		TotalAmount: 60.00,
		PaidBy:      charlie.User.Id,
		SplitType:   expensepb.SplitType_EXACT,
		Splits: []*expensepb.Split{
			{UserId: alice.User.Id, Amount: 25.00},
			{UserId: bob.User.Id, Amount: 15.00},
			{UserId: charlie.User.Id, Amount: 20.00},
		},
	})
	if err != nil {
		log.Fatalf("failed to create expense: %v", err)
	}
	fmt.Printf("   ✓ Expense created: %s ($%.2f)\n", expense3.Expense.Description, expense3.Expense.TotalAmount)
	fmt.Printf("     Paid by Charlie, custom split\n\n")

	// 4. Check balances
	fmt.Println("4. Checking balances...")

	aliceBalances, err := expenseClient.GetBalances(ctx, &expensepb.GetBalancesRequest{
		UserId: alice.User.Id,
	})
	if err != nil {
		log.Fatalf("failed to get balances: %v", err)
	}
	fmt.Printf("   Alice's balances:\n")
	for _, balance := range aliceBalances.Balances {
		if balance.FromUserId == alice.User.Id {
			fmt.Printf("     → Owes $%.2f to user %s\n", balance.Amount, balance.ToUserId)
		} else {
			fmt.Printf("     ← Is owed $%.2f from user %s\n", balance.Amount, balance.FromUserId)
		}
	}

	bobBalances, err := expenseClient.GetBalances(ctx, &expensepb.GetBalancesRequest{
		UserId: bob.User.Id,
	})
	if err != nil {
		log.Fatalf("failed to get balances: %v", err)
	}
	fmt.Printf("   Bob's balances:\n")
	for _, balance := range bobBalances.Balances {
		if balance.FromUserId == bob.User.Id {
			fmt.Printf("     → Owes $%.2f to user %s\n", balance.Amount, balance.ToUserId)
		} else {
			fmt.Printf("     ← Is owed $%.2f from user %s\n", balance.Amount, balance.FromUserId)
		}
	}

	charlieBalances, err := expenseClient.GetBalances(ctx, &expensepb.GetBalancesRequest{
		UserId: charlie.User.Id,
	})
	if err != nil {
		log.Fatalf("failed to get balances: %v", err)
	}
	fmt.Printf("   Charlie's balances:\n")
	for _, balance := range charlieBalances.Balances {
		if balance.FromUserId == charlie.User.Id {
			fmt.Printf("     → Owes $%.2f to user %s\n", balance.Amount, balance.ToUserId)
		} else {
			fmt.Printf("     ← Is owed $%.2f from user %s\n", balance.Amount, balance.FromUserId)
		}
	}
	fmt.Println()

	// 5. Settle a balance
	fmt.Println("5. Settling balances...")
	settle, err := expenseClient.SettleBalance(ctx, &expensepb.SettleBalanceRequest{
		FromUserId: bob.User.Id,
		ToUserId:   alice.User.Id,
		Amount:     30.00,
	})
	if err != nil {
		log.Fatalf("failed to settle balance: %v", err)
	}
	fmt.Printf("   ✓ %s\n", settle.Message)
	if settle.RemainingBalance != nil {
		fmt.Printf("     Remaining balance: $%.2f\n\n", settle.RemainingBalance.Amount)
	} else {
		fmt.Printf("     Remaining balance: $0.00\n\n")
	}

	// 6. List expenses
	fmt.Println("6. Listing Alice's expenses...")
	aliceExpenses, err := expenseClient.ListExpenses(ctx, &expensepb.ListExpensesRequest{
		UserId: alice.User.Id,
	})
	if err != nil {
		log.Fatalf("failed to list expenses: %v", err)
	}
	for _, exp := range aliceExpenses.Expenses {
		fmt.Printf("   - %s: $%.2f (paid by %s)\n", exp.Description, exp.TotalAmount, exp.PaidBy)
	}

	fmt.Println("\n=== Demo completed successfully! ===")
}
