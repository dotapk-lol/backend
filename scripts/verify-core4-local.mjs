// Local candidate protocol smoke. Synthetic results, not skill/browser acceptance.
import assert from 'node:assert/strict';
import fs from 'node:fs/promises';
import path from 'node:path';
import {pathToFileURL} from 'node:url';
import {randomUUID,createHash} from 'node:crypto';
const frontend=process.argv[2];if(!frontend)throw Error('Pass CORE4 worktree');
const {MatchAPI,resultPayload}=await import(pathToFileURL(path.join(frontend,'src/match-api.js')));
const {NET_VERSION}=await import(pathToFileURL(path.join(frontend,'src/net-version.js')));
assert.equal(NET_VERSION,'duel-751bcab20194934a863a');
const base='http://127.0.0.1:18083/api/v1',origin='http://127.0.0.1:4174',rosterId='arena-core4-24-v1';
const fetcher=async(url,options={})=>{const r=await fetch(url,{...options,headers:{...options.headers,Origin:origin}});assert.equal(r.headers.get('access-control-allow-origin'),origin);return r;};
const api=new MatchAPI({base,fetcher,storage:null}),peer=new MatchAPI({base,fetcher,storage:null});
const checks=[],receipts=[],key=()=>randomUUID(),pass=s=>checks.push(s);
const rejects=async(code,fn)=>assert.rejects(fn,e=>e.status===code);
const meta=await api.loadRegistry();assert.equal(meta.status,'verified');assert.equal(meta.rosterId,rosterId);assert.deepEqual(meta.heroIds,[...Array(20).keys(),25,31,45,100]);pass('actual CORE4 adapter validates exact24 candidate roster');
const result=resultPayload({history:[{winner:0,remaining:1.2},{winner:0,remaining:2}]});
const resultHash=createHash('sha256').update(JSON.stringify(result)).digest('hex');
for(const [hero,opponentHero] of [[25,31],[31,45],[45,100],[100,25]]){
 const requestId=key(),m=await api.createPVE(hero,opponentHero,requestId);assert.equal(m.rosterId,rosterId);assert.equal(m.version,NET_VERSION);assert.equal((await api.createPVE(hero,opponentHero,requestId)).id,m.id);
 assert.equal((await api.submit(m.id,result)).status,'recorded');receipts.push({kind:'core4-pve',id:m.id,heroes:[hero,opponentHero],expectedStatus:'recorded',resultDigest:resultHash});
}pass('all four stable IDs accepted, snapshotted and idempotently recorded');
for(const [transport,heroes] of [['local',[25,100]],['broadcastchannel',[31,45]]]){
 const m=await api.createLocal(...heroes,transport,key());assert.equal((await api.submit(m.id,result)).status,'recorded');receipts.push({kind:transport,id:m.id,heroes,expectedStatus:'recorded',resultDigest:resultHash});
}pass('local and BC keep client_reported semantics with new IDs');
const policy={direction:'above',rttMs:200,jitterMs:30,lossPct:5,minSamples:24,window:30,maxAgeMs:3000};
const roomBody={version:NET_VERSION,hero:25,offer:{type:'offer',sdp:'v=0\r\n'},policy,rosterId};
await rejects(400,()=>api.request('rooms',{...roomBody,requestId:key(),rosterId:'unknown-roster'}));
await rejects(409,()=>api.request('rooms',{...roomBody,requestId:key(),version:'duel-unbound-core4'}));
pass('room creation rejects unknown roster and unbound build');
const room=await api.request('rooms',{...roomBody,requestId:key()});
for(const [override,code] of [[{rosterId:'unknown-roster'},400],[{version:'duel-unbound-core4'},409],[{rosterId:undefined},409],[{rosterId:'legacy-20-v1'},409],[{hero:20},400]]){
 await rejects(code,()=>peer.request('rooms/join',{code:room.code,version:NET_VERSION,hero:31,rosterId,...override}));
}pass('join rejects unknown roster/build, omitted/legacy fallback and unplayable hero');
await peer.request('rooms/join',{code:room.code,version:NET_VERSION,hero:31,rosterId});
await peer.request('rooms/'+room.id+'/answer',{version:NET_VERSION,answer:{type:'answer',sdp:'v=0\r\n'}});
const m=await api.createMatch(room.id,key());assert.equal(m.rosterId,rosterId);
await api.ready(m.id);await peer.ready(m.id);
assert.equal((await api.submit(m.id,result)).status,'pending');assert.equal((await peer.submit(m.id,result)).status,'confirmed');
receipts.push({kind:'core4-pvp',id:m.id,heroes:[25,31],expectedStatus:'confirmed',resultDigest:resultHash});pass('new IDs retain two-party confirmed lifecycle');
// Legacy wire compatibility must use an old build, never the bound CORE4 build.
const oldBody={requestId:key(),version:'duel-legacy-core4-check',hero:0,opponentHero:3,aiDifficulty:'normal'};
const old=await api.request('matches/pve',oldBody);assert.equal(old.rosterId,'legacy-20-v1');assert.equal((await api.request('matches/pve',oldBody)).id,old.id);
const oldResult={...result,version:oldBody.version};assert.equal((await api.submit(old.id,oldResult)).status,'recorded');
receipts.push({kind:'legacy-pve-omitted',id:old.id,heroes:[0,3],expectedStatus:'recorded',createDigest:createHash('sha256').update(JSON.stringify(oldBody)).digest('hex'),resultDigest:createHash('sha256').update(JSON.stringify(oldResult)).digest('hex')});
for(const wrong of [{version:NET_VERSION},{version:NET_VERSION,rosterId:'legacy-20-v1'},{version:NET_VERSION,rosterId,hero:20}]){
 await rejects(wrong.hero===20?400:409,()=>api.request('matches/pve',{...oldBody,requestId:key(),...wrong}));
}pass('old omitted roster retains original hashes; bound CORE4 build cannot fall back');
const legacyRoom=await api.request('rooms',{requestId:key(),version:oldBody.version,hero:0,offer:{type:'offer',sdp:'v=0\r\n'},policy});
assert.equal((await peer.request('rooms/join',{code:legacyRoom.code,version:oldBody.version,hero:3})).rosterId,'legacy-20-v1');
await api.request('rooms/'+legacyRoom.id,undefined,'DELETE');pass('old omitted-roster room creation and join still work');
const health=await(await fetch('http://127.0.0.1:18082/healthz')).json();assert.equal(health.contractVersion,'v1.2-abort-reconciliation');pass('18082 still v1.2');
const evidence={checkedAt:new Date().toISOString(),scope:'Local candidate only; actual frontend modules and HTTP; synthetic results/SDP, no skill or browser acceptance',runtime:NET_VERSION,rosterId,base,corsOrigin:origin,checks,receipts,productionChanged:false};
await fs.writeFile(new URL('../docs/core4-local-http-evidence.json',import.meta.url),JSON.stringify(evidence,null,2)+'\n');console.log(JSON.stringify({passed:checks.length,matches:receipts.length,base,runtime:NET_VERSION}));
