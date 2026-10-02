#!/usr/bin/env python3
"""Local/BC client-reported PVP smoke. Local endpoint only; prints no tokens."""
import json,os,uuid,urllib.request,urllib.error
base=os.environ.get('DUEL_API_URL','http://127.0.0.1:18083')
if not base.startswith(('http://127.0.0.1:', 'http://localhost:')): raise SystemExit('Local QA only')
def call(method,path,body=None,token=None,want=200):
    headers={'Content-Type':'application/json'}
    if token: headers['Authorization']='Bearer '+token
    req=urllib.request.Request(base+path,data=None if body is None else json.dumps(body).encode(),headers=headers,method=method)
    try:
        with urllib.request.urlopen(req,timeout=10) as r: status,data=r.status,json.load(r)
    except urllib.error.HTTPError as e: status,data=e.code,json.load(e)
    assert status==want,(path,status,data)
    return data
health=call('GET','/healthz')
assert health.get('contractVersion') in ('v1.1-local-pvp','v1.2-abort-reconciliation'),health
owner=call('POST','/api/v1/sessions',{},want=201)
other=call('POST','/api/v1/sessions',{},want=201)
version='qa-local-pvp-v2'
report={'version':version,'outcome':'completed','rounds':[{'number':1,'winner':1,'remainingMs':0},{'number':2,'winner':1,'remainingMs':1000}],'score':[0,2],'winner':1,'reason':''}
records=[]
for transport in ['local','broadcastchannel']:
    body={'requestId':uuid.uuid4().hex,'version':version,'hero':2,'opponentHero':5,'transport':transport}
    m=call('POST','/api/v1/matches/local',body,owner['token'],201)
    duplicate=call('POST','/api/v1/matches/local',body,owner['token'],201)
    assert m['id']==duplicate['id'] and m['mode']=='pvp' and m['trust']=='client_reported'
    assert m['participantKinds']==['local_slot','local_slot'] and m['reporterPlayerId']==owner['playerId']
    assert m['players'][0]['id']!=m['players'][1]['id'] and all(p['id']!=owner['playerId'] for p in m['players'])
    call('GET','/api/v1/matches/'+m['id'],token=other['token'],want=403)
    call('POST','/api/v1/matches/local',{**body,'trust':'peer_agreement'},owner['token'],400)
    result=call('POST','/api/v1/matches/'+m['id']+'/results',report,owner['token'])
    assert result['status']=='recorded' and result['reported']==[True,False]
    duplicate=call('POST','/api/v1/matches/'+m['id']+'/results',report,owner['token'])
    assert duplicate['status']=='recorded' and duplicate['id']==m['id']
    records.append({'id':m['id'],'mode':result['mode'],'transport':transport,'trust':result['trust'],'status':result['status'],'score':result['score'],'reported':result['reported']})
print(json.dumps({'health':health,'records':records,'createAndSubmitIdempotent':True,'outsiderRejected':True,'trustPromotionRejected':True},indent=2))
