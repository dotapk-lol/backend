#!/usr/bin/env python3
"""Creates marked test games against an explicitly selected local QA endpoint; prints no tokens."""
import json, os, uuid, urllib.request, urllib.error
base=os.environ.get('DUEL_API_URL','http://127.0.0.1:18082')
if not base.startswith(('http://127.0.0.1:', 'http://localhost:')):
    raise SystemExit('Smoke script is local QA only')
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
a=call('POST','/api/v1/sessions',{},want=201)
b=call('POST','/api/v1/sessions',{},want=201)
version='qa-smoke-go-mysql-v1'
policy={'direction':'above','rttMs':200,'jitterMs':30,'lossPct':5,'minSamples':24,'window':30,'maxAgeMs':3000}
room=call('POST','/api/v1/rooms',{'requestId':uuid.uuid4().hex,'version':version,'hero':0,'offer':{'type':'offer','sdp':'v=0\r\n'},'policy':policy},a['token'],201)
assert len(room['code'])==6 and room['code'].isdigit()
call('POST','/api/v1/rooms/join',{'code':room['code'],'version':version,'hero':3},b['token'])
call('POST','/api/v1/rooms/'+room['id']+'/answer',{'version':version,'answer':{'type':'answer','sdp':'v=0\r\n'}},b['token'])
m=call('POST','/api/v1/rooms/'+room['id']+'/matches',{'requestId':uuid.uuid4().hex,'version':version},a['token'],201)
call('POST','/api/v1/matches/'+m['id']+'/ready',{'version':version},a['token'])
started=call('POST','/api/v1/matches/'+m['id']+'/ready',{'version':version},b['token'])
assert started['status']=='in_progress'
report={'version':version,'outcome':'completed','rounds':[{'number':1,'winner':0,'remainingMs':500},{'number':2,'winner':0,'remainingMs':300}],'score':[2,0],'winner':0,'reason':''}
malformed={**report,'score':[2,0,0]}
call('POST','/api/v1/matches/'+m['id']+'/results',malformed,a['token'],400)
pending=call('POST','/api/v1/matches/'+m['id']+'/results',report,a['token'])
assert pending['status']=='pending' and 'submissions' not in pending
confirmed=call('POST','/api/v1/matches/'+m['id']+'/results',report,b['token'])
assert confirmed['status']=='confirmed'
replay=call('POST','/api/v1/matches/'+m['id']+'/results',report,a['token'])
assert replay['status']=='confirmed' and replay['id']==m['id']
pve=call('POST','/api/v1/matches/pve',{'requestId':uuid.uuid4().hex,'version':version,'hero':1,'opponentHero':2,'aiDifficulty':'normal'},a['token'],201)
recorded=call('POST','/api/v1/matches/'+pve['id']+'/results',report,a['token'])
assert recorded['status']=='recorded' and recorded['trust']=='client_reported'
print(json.dumps({'health':health,'version':version,'pvpMatchId':m['id'],'pvpStatus':confirmed['status'],'pvpReported':confirmed['reported'],'pveMatchId':pve['id'],'pveStatus':recorded['status'],'pveTrust':recorded['trust'],'strictScoreLength':True,'duplicateReportIdempotent':True,'peerReportHidden':True},ensure_ascii=False,indent=2))
