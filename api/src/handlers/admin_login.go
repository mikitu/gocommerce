package handlers

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/labstack/echo"
	apihttp "github.com/mikitu/gocommerce/api/src/http"
	"github.com/spf13/viper"
	"net/http"
)

func AdminLoginHandler(c echo.Context) error {
	aws_region := viper.GetString("aws_region")
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(aws_region),
	}))
	client := cognitoidentityprovider.New(sess);
	params := &cognitoidentityprovider.AdminInitiateAuthInput{
		AuthFlow: aws.String("ADMIN_NO_SRP_AUTH"),
		AuthParameters: map[string]*string{
			"USERNAME": aws.String(c.FormValue("username")),
			"PASSWORD": aws.String(c.FormValue("password")),
		},
		ClientId: aws.String(viper.GetString("cognito_client_id")),
		UserPoolId: aws.String(viper.GetString("userPoolId")),
	}
	resp, err := client.AdminInitiateAuth(params)
	if err != nil {
		return c.JSON(http.StatusBadRequest, &apihttp.ResponseFormatter{Status: http.StatusBadRequest, Data: nil, Errors: err})
	}
	return c.JSON(http.StatusOK, &apihttp.ResponseFormatter{Status: http.StatusOK, Data: resp, Errors: nil})
}
