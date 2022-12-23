package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/plaid/plaid-go/v10/plaid"
)

type AccountBase struct {
	AccountId   string `json:"accountId"`
	Name        string `json:"name"`
	ProductName string `json:"productName"`
	Type        string `json:"type"`
	Balance     string `json:"balance"`
}

type CreditAccount struct {
	Account              AccountBase `json:"account"`
	NextPaymentDate      string      `json:"nextPaymentDate"`
	MinimumPaymentAmount float64     `json:"minimumPaymentAmount"`
}

func GetClient() *plaid.APIClient {
	configuration := plaid.NewConfiguration()
	configuration.AddDefaultHeader(PLAID_CLIENT_ID, os.Getenv("PLAID_CLIENT_ID"))
	configuration.AddDefaultHeader(PLAID_SECRET, os.Getenv("PLAID_SECRET"))
	configuration.UseEnvironment(plaid.Sandbox)

	return plaid.NewAPIClient(configuration)
}

func getAccessToken(client *plaid.APIClient, ctx context.Context) (string, error) {
	publicTokenResp, resp, err := client.PlaidApi.SandboxPublicTokenCreate(ctx).SandboxPublicTokenCreateRequest(
		*plaid.NewSandboxPublicTokenCreateRequest(
			"ins_109508",
			[]plaid.Products{plaid.PRODUCTS_ASSETS},
		),
	).Execute()

	fmt.Printf("%s\n", resp.Body)
	if err != nil {
		return "", fmt.Errorf("error in getting public token: %q", err)
	}

	exchangePublicTokenResp, _, err := client.PlaidApi.ItemPublicTokenExchange(ctx).ItemPublicTokenExchangeRequest(
		*plaid.NewItemPublicTokenExchangeRequest(publicTokenResp.GetPublicToken()),
	).Execute()

	if err != nil {
		return "", fmt.Errorf("error in getting access token: %q", err)
	}

	return exchangePublicTokenResp.GetAccessToken(), nil
}

func getCreditAccBalance(client *plaid.APIClient, ctx context.Context) ([]string, error) {
	accessToken, err := getAccessToken(client, ctx)

	if err != nil {
		return nil, fmt.Errorf("error getting balance: %q", err)
	}

	balancesGetReq := plaid.NewAccountsBalanceGetRequest(accessToken)
	balancesGetResp, _, err := client.PlaidApi.AccountsBalanceGet(ctx).AccountsBalanceGetRequest(
		*balancesGetReq,
	).Execute()

	if err != nil {
		return nil, fmt.Errorf("error getting balance: %q", err)
	}

	var creditAccs []string

	for _, acc := range balancesGetResp.Accounts {
		if acc.Type == plaid.ACCOUNTTYPE_CREDIT {
			creditAccs = append(creditAccs, acc.GetAccountId())
		}
		fmt.Println(acc.AccountId)
		fmt.Printf("Name: %s, Type: %s, Balances: %v%v, Limit: %v\n", acc.Name, acc.Type, acc.Balances.GetCurrent(), acc.Balances.GetIsoCurrencyCode(), acc.Balances.GetLimit())
	}

	return creditAccs, nil
}

func SearchInstitutions(client *plaid.APIClient, ctx context.Context, query string) ([]plaid.Institution, error) {

	request := plaid.NewInstitutionsSearchRequest(query, []plaid.Products{plaid.PRODUCTS_ASSETS}, []plaid.CountryCode{plaid.COUNTRYCODE_CA, plaid.COUNTRYCODE_US})
	response, _, err := client.PlaidApi.InstitutionsSearch(ctx).InstitutionsSearchRequest(*request).Execute()

	if err != nil {
		return nil, fmt.Errorf("error searching: %q", err)
	}

	return response.Institutions, nil
}

func GetLiabilities(client *plaid.APIClient, ctx context.Context, accountIds []string) ([]CreditAccount, error) {
	accessToken, err := getAccessToken(client, ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting balance: %q", err)
	}

	request := plaid.NewLiabilitiesGetRequest(accessToken)
	request.SetOptions(*plaid.NewLiabilitiesGetRequestOptions())
	// request.Options.SetAccountIds(accountIds)

	log.Println("Fetching liabilities for:", accountIds)
	resp, _, err := client.PlaidApi.LiabilitiesGet(ctx).LiabilitiesGetRequest(*request).Execute()

	// fmt.Println(httpResp)

	if err != nil {
		return nil, fmt.Errorf("error getting liability: %q", err)
	}

	var liabilities []CreditAccount

	for _, acc := range resp.Accounts {
		if acc.Type == plaid.ACCOUNTTYPE_CREDIT {
			fmt.Printf("Name: %s, Type: %s, Balances: %v%v, Limit: %v\n", acc.Name, acc.Type, acc.Balances.GetCurrent(), acc.Balances.GetIsoCurrencyCode(), acc.Balances.GetLimit())
			account := &AccountBase{AccountId: acc.AccountId, ProductName: acc.GetOfficialName(), Name: acc.Name, Type: string(acc.GetSubtype()), Balance: fmt.Sprintf("%.2f%s", acc.Balances.GetCurrent(), acc.Balances.GetIsoCurrencyCode())}
			for _, liab := range resp.GetLiabilities().Credit {
				if liab.GetAccountId() == acc.AccountId {
					creditAccount := &CreditAccount{Account: *account, NextPaymentDate: liab.GetNextPaymentDueDate(), MinimumPaymentAmount: liab.GetLastPaymentAmount()}
					liabilities = append(liabilities, *creditAccount)
					fmt.Printf("Next payment date: %v, Min amount to pay: %v\n", liab.GetNextPaymentDueDate(), liab.GetMinimumPaymentAmount())
				}
			}
		}
	}

	return liabilities, nil
}
