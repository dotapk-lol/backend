#!/usr/bin/env python3
"""Local HTTP abort-reconciliation regression. No tokens printed or retained."""
import json,os,uuid,urllib.request
base=os.environ.get('DUEL_API_URL','http://127.0.0.1:18082')
if not base.startswith(('http://127.0.0.1:', 'http://localhost:')): raise SystemExit('Local QA only')
def call(path,body=None,token=None,method='POST'):
    headers={'Content-Type':'application/json'}
    if token: headers['Authorization']='Bearer '+token
    req=urllib.request.Request(base+path,data=None if body is None else json.dumps(body).encode(),headers=headers,method=method)
    with urllib.request.urlopen(req,timeout=10) as r:return json.load(r)
health=call('/healthz',method='GET')
assert health['contractVersion']=='v1.2-abort-reconciliation'
a=call('/api/v1/sessions',{})
b=call('/api/v1/sessions',{})
version='qa-abort-v12'
policy={'direction':'above','rttMs':200,'jitterMs':30,'lossPct':5,'minSamples':24,'window':30,'maxAgeMs':3000}
r=call('/api/v1/rooms',{'requestId':uuid.uuid4().hex,'version':version,'hero':0,'offer':{'type':'offer','sdp':'v=0\r\n'},'policy':policy},a['token'])
call('/api/v1/rooms/join',{'code':r['code'],'version':version,'hero':1},b['token'])
call('/api/v1/rooms/'+r['id']+'/answer',{'version':version,'answer':{'type':'answer','sdp':'v=0\r\n'}},b['token'])
def abort(reason):return {'version':version,'outcome':'aborted','rounds':[],'score':[0,0],'winner':-1,'reason':reason}
def win(side):return {'version':version,'outcome':'completed','rounds':[{'number':i+1,'winner':side,'remainingMs':0} for i in range(2)],'score':[2 if i==side else 0 for i in range(2)],'winner':side,'reason':''}
outputs=[]
for name,reports,status,reason in [('normal_disconnect',[abort('left'),abort('disconnect')],'aborted','interrupted'),('completed_abort_conflict',[win(0),abort('disconnect')],'disputed','outcome_conflict'),('completed_winner_conflict',[win(0),win(1)],'disputed','conflicting_reports'),('completed_agreement',[win(0),win(0)],'confirmed','')]:
    m=call('/api/v1/rooms/'+r['id']+'/matches',{'requestId':uuid.uuid4().hex,'version':version},a['token'])
    path='/api/v1/matches/'+m['id']
    call(path+'/ready',{'version':version},a['token']);call(path+'/ready',{'version':version},b['token'])
    first=call(path+'/results',reports[0],a['token']);assert first['status']=='pending'
    final=call(path+'/results',reports[1],b['token'])
    assert final['status']==status and final['reason']==reason,final
    duplicate=call(path+'/results',reports[0],a['token']);assert duplicate==final
    outputs.append({'case':name,'id':final['id'],'status':final['status'],'reason':final['reason'],'scoreAgreement':final['scoreAgreement'],'winner':final['winner'],'reported':final['reported']})
print(json.dumps({'health':health,'cases':outputs,'exactRetriesImmutable':True},indent=2))
