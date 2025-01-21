# VIAC 3a Wealth Getter API
API to get the wealth of a VIAC 3a account (based on [this](https://github.com/ThisIsntTheWay/viac-wealth-getter)).  
I use this in tandem with a Google sheet.

## Usage
```bash
cp .env.example .env
docker compose up -d

curl http://localhost:8080/wealth
```

### Response
```json
{
    "timestamp": "2025-01-21T21:37:52.082522743+01:00",
    "wealth": {
        "totalValue": 123,
        "totalPerformance": 0.123,
        "totalReturn": 123
    },
    "upToDate": true,
    "error": ""
}
```

> [!WARNING]  
> If `upToDate` is `false`, then a **cached** entry will be shown.  
> In such case, the field `error` will be populated and `timestamp` will be the time of the last successful update.  
> Additionally, only `.wealth.totalValue` will be populated.
