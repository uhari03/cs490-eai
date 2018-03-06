import json
import requests
import random

if __name__ == '__main__':

    eai_base = "https://lit-sea-20426.herokuapp.com"
    dummy_base = "https://stil-taiga-23440.herokuapp.com"

    register_system = '/register/system'
    register_topic = '/register/topic'
    subscribe = '/subscribe'
    publish = '/publish'

    salt = str(random.randint(10, 10000))

    sales_system = {'systemName': 'sales' + salt, 'applicationEndpoint': 'localhost'}
    ais_name = 'ais' + salt
    ais_system = {'systemName': ais_name, 'applicationEndpoint': dummy_base}

    structure = {
      "cust_id": "(int) ID of the purchaser",
      "amt": "(float) Amount of the sale in CAD",
      "dt": "(datetime) Time and date of the sale"
    }

    topic_name = 'new_sale' + salt
    # ugly hack: right now, Go wants structure to be a string representing
    # valid JSON instead of just valid JSON itself
    topic_data = {
       'topicName': topic_name,
       'description': 'Fires whenever a new sale is made',
       'owner': 'sales',
       'structure':
       '''
       {
          "cust_id": "(int) ID of the purchaser",
          "amt": "(float) Amount of the sale in CAD",
          "dt": "(datetime) Time and date of the sale"
       }
       '''
    }

    topic_print = {
       'topicName': topic_name,
       'description': 'Fires whenever a new sale is made',
       'owner': 'sales',
       'structure': structure
    }

    subscription = {'systemName': ais_name, 'topicName': topic_name}

    event_data = '''{"cust_id": 1337, "amt": 314159.27, "dt": "2018-04-13"}'''
    # '{"topicName": "testtopic1", "data": "{\"struct1\": \"value1\"}"}'
    event = {'topicName': topic_name, 'data': event_data}

    print('~~~ Step 1: register internal systems (sales/AIS) ~~~')
    print('Sales system')
    print(json.dumps(sales_system, indent = 4))
    print('AIS system')
    print(json.dumps(ais_system, indent = 4))

    r = requests.post(eai_base + register_system, json = sales_system)
    print('Status for sales registration: {}'.format(r.status_code))
    r = requests.post(eai_base + register_system, json = ais_system)
    print('Status for AIS registration: {}'.format(r.status_code))

    input('~~~ Step 2: create a topic ~~~')
    print(json.dumps(topic_print, indent = 4))

    r = requests.post(eai_base + register_topic, json = topic_data)
    print('Status for topic registration: {}'.format(r.status_code))
    print('Response from topic registration: {}'.format(r.text))

    input('~~~ Step 3: subscribe AIS system to sales event ~~~')
    print(json.dumps(subscription, indent = 4))
    r = requests.get(eai_base + subscribe, params = subscription)
    print('Status for event subscription: {}'.format(r.status_code))

    # loops until you enter "stop", sending one event each time
    input('~~~ Step 4: send events (type "stop" to end) ~~~')
    while True:
        command = input()
        if command == 'stop':
            break

        r = requests.post(eai_base + publish, json = event)
        print('Sent a new {} event; saw status {}'.format(topic_name, r.status_code))
