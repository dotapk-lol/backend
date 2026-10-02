#!/usr/bin/env python3
"""Authorized DotaPK HTTPS acceptance. No tokens are printed or persisted."""
import json,urllib.request,urllib.error,uuid
base='https://api.dotapk.lol'; origin='https://dotapk.lol'
def request(method,path,body=None,token=None,origin_value=origin,expected=200,extra=None):
    headers={'Content-Type':'application/json','Origin':origin_value}
    if token:headers['Authorization']='Bearer '+token
    if extra:headers.update(extra)
    req=urllib.request.Request(base+path,data=None if body is None else json.dumps(body).encode(),headers=headers,method=method)
    try:
        with urllib.request.urlopen(req,timeout=15) as r:
            raw=r.read();status=r.status;response_headers=dict(r.headers.items())
    except urllib.error.HTTPError as e:raw=e.read();status=e.code;response_headers=dict(e.headers.items())
    if status!=expected:raise RuntimeError('Unexpected status for '+path+': '+str(status))
    data=json.loads(raw) if raw else None
    lower={k.lower():v for k,v in response_headers.items()}
    return data,lower
health,h=request('GET','/healthz')
assert health['ok'] and health['contractVersion']=='v1.2-abort-reconciliation'
assert h.get('access-control-allow-origin')==origin
_,pre=request('OPTIONS','/api/v1/matches/pve',expected=204,extra={'Access-Control-Request-Method':'POST','Access-Control-Request-Headers':'content-type,authorization'})
assert pre.get('access-control-allow-origin')==origin
_,denied=request('GET','/healthz',origin_value='https://untrusted.example',expected=403)
assert 'access-control-allow-origin' not in denied
owner,_=request('POST','/api/v1/sessions',{},expected=201)
version='qa-public-https-v12'
m,_=request('POST','/api/v1/matches/pve',{'requestId':uuid.uuid4().hex,'version':version,'hero':0,'opponentHero':3,'aiDifficulty':'normal'},owner['token'],expected=201)
report={'version':version,'outcome':'completed','rounds':[{'number':1,'winner':0,'remainingMs':1000},{'number':2,'winner':0,'remainingMs':500}],'score':[2,0],'winner':0,'reason':''}
final,_=request('POST','/api/v1/matches/'+m['id']+'/results',report,owner['token'])
assert final['status']=='recorded' and final['trust']=='client_reported'
replay,_=request('POST','/api/v1/matches/'+m['id']+'/results',report,owner['token'])
assert replay==final
fresh,_=request('GET','/api/v1/matches/'+m['id'],token=owner['token'])
assert fresh==final
print(json.dumps({'base':base,'origin':origin,'health':health,'corsExactOrigin':True,'preflightPassed':True,'foreignOriginRejected':True,'httpsCertificateVerification':'system trust and hostname validation enabled','matchId':m['id'],'mode':final['mode'],'status':final['status'],'trust':final['trust'],'score':final['score'],'idempotentReplay':True,'readAfterWrite':True},indent=2))
