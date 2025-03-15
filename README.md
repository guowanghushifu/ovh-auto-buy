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
            - APP_KEY=your_ovh_app_key
            - APP_SECRET=your_ovh_app_secret
            - CONSUMER_KEY=your_ovh_consumer_key
            - REGION=ovh-eu
            - ZONE=IE
            - TG_TOKEN=your_telegram_token
            - TG_CHATID=your_telegram_chat_id
            - DATACENTER=sbg,rbx,fra
            - PLANCODE=24sk20
            - FQN=24sk20.ram-32g-ecc-2133.softraid-2x450nvme
            - OPTIONS=bandwidth-300-24sk,ram-32g-ecc-2133-24sk20,softraid-2x450nvme-24sk20
            - AUTOPAY=true
            - FREQUENCY=30
            - BUYNUM=1
            - TZ=Asia/Shanghai
        image: guowanghushifu/ovh-auto-buy:latest
```
Some Env Explain:
- BUYNUM: How many servers to buy, e.g. 2
- DATACENTER: Which Datacenter to buy, e.g. `"sbg,rbx,fra"`. It should be a single value or comma-separated list. If you don't need to filter datacenter, `set it to "any"`
- FREQUENCY: How many seconds to sleep between each try, e.g. 30
- FQN: The detailed configration of your PLANCODE. e.g. `"24sk20.ram-32g-ecc-2133.softraid-2x450nvme"`. It should be a single value or comma-separated list. If you don't need to filter FQN, set it to `"any"`. If you don't know what to set, tpye `"hoho"`, run the docker once, find FQN list at the **HEAD** of log, then modify this to the right one.
- AUTOPAY: Auto pay bill, e.g. true

