# Automatically buys OVH Eco Servers (Docker)
Go source code modified from: https://blog.yessure.org/index.php/archives/203

To generate token:
- IE: `https://eu.api.ovh.com/createToken/index.cgi?GET=/*&PUT=/*&POST=/*&DELETE=/*`
- CA: `https://ca.api.ovh.com/createToken/index.cgi?GET=/*&PUT=/*&POST=/*&DELETE=/*`

To get Eco planCode and options:
- IE: https://eu.api.ovh.com/v1/order/catalog/public/eco?ovhSubsidiary=IE
- CA: https://ca.api.ovh.com/v1/order/catalog/public/eco?ovhSubsidiary=CA

## docker-compose.yml
```yaml
services:
    ovh:
        container_name: ovh
        environment:
            - APP_KEY=123456
            - APP_SECRET=123456
            - CONSUMER_KEY=123456
            - REGION=ovh-eu
            - ZONE=IE
            - TG_TOKEN=123456:123456
            - TG_CHATID=-100123456
            - DATACENTER=sbg,rbx,fra
            - PLANCODE=24sk20
            - FQN=24sk20.ram-32g-ecc-2133.softraid-2x450nvme
            - OPTIONS=bandwidth-300-24sk,ram-32g-ecc-2133-24sk20,softraid-2x450nvme-24sk20
            - AUTOPAY=true
            - FREQUENCY=15
            - BUYNUM=1
        image: guowanghushifu/ovh-auto-buy:latest
```
Some Env Para Explain:
- BUYNUM: How many servers to buy, eg. 2
- DATACENTER: Which Datacenter to buy, eg. sbg,rbx,fra
- FREQUENCY: How many seconds to sleep between each try
- AUTOPAY: Auto pay money?

