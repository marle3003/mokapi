import{M as b,r as a,c as U,j as e,T as D,t as R,W as M,a as C,b as F}from"./assets/wsPort--EbOV_hL.js";const O=()=>{const[l,k]=a.useState(!1),[c,d]=a.useState([]),[u,E]=a.useState([]),[p,j]=a.useState(N),[v,m]=a.useState({done:0,total:0}),[T,h]=a.useState(!1),[y,f]=a.useState(null),[w,L]=a.useState(null),g=a.useCallback(r=>{const o=[],t=[],s=new URL(window.location.href);for(let i=0;i<r.length;i++){const n=r.item(i);if(!n)continue;const S=URL.createObjectURL(n);o.push(S),t.push(n.name),s.searchParams.append("trace",S),s.searchParams.append("traceFileName",n.name)}const x=s.toString();window.history.pushState({},"",x),d(o),E(t),h(!1),f(null)},[]),P=a.useCallback(r=>{r.preventDefault(),g(r.dataTransfer.files)},[g]),W=a.useCallback(r=>{r.preventDefault(),r.target.files&&g(r.target.files)},[g]);return a.useEffect(()=>{const r=new URL(window.location.href).searchParams,o=r.getAll("trace");k(r.has("isServer"));for(const t of o)if(t.startsWith("file:")){L(t||null);return}r.has("isServer")?U({onEvent(t,s){t==="loadTrace"&&(d(s.url?[s.url]:[]),h(!1),f(null))},onClose(){}}).then(t=>{t("ready")}):o.some(t=>t.startsWith("blob:"))||d(o)},[]),a.useEffect(()=>{(async()=>{if(c.length){const r=s=>{s.data.method==="progress"&&m(s.data.params)};navigator.serviceWorker.addEventListener("message",r),m({done:0,total:1});const o=[];for(let s=0;s<c.length;s++){const x=c[s],i=new URLSearchParams;i.set("trace",x),u.length&&i.set("traceFileName",u[s]);const n=await fetch(`contexts?${i.toString()}`);if(!n.ok){l||d([]),f((await n.json()).error);return}o.push(...await n.json())}navigator.serviceWorker.removeEventListener("message",r);const t=new b(o);m({done:0,total:0}),j(t)}else j(N)})()},[l,c,u]),e.jsxs("div",{className:"vbox workbench-loader",onDragOver:r=>{r.preventDefault(),h(!0)},children:[e.jsxs("div",{className:"hbox header",children:[e.jsx("div",{className:"logo",children:e.jsx("img",{src:"playwright-logo.svg",alt:"Playwright logo"})}),e.jsx("div",{className:"product",children:"Playwright"}),p.title&&e.jsx("div",{className:"title",children:p.title}),e.jsx("div",{className:"spacer"}),e.jsx(D,{icon:"color-mode",title:"Toggle color mode",toggled:!1,onClick:()=>R()})]}),e.jsx("div",{className:"progress",children:e.jsx("div",{className:"inner-progress",style:{width:v.total?100*v.done/v.total+"%":0}})}),e.jsx(M,{model:p}),w&&e.jsxs("div",{className:"drop-target",children:[e.jsx("div",{children:"Trace Viewer uses Service Workers to show traces. To view trace:"}),e.jsxs("div",{style:{paddingTop:20},children:[e.jsxs("div",{children:["1. Click ",e.jsx("a",{href:w,children:"here"})," to put your trace into the download shelf"]}),e.jsxs("div",{children:["2. Go to ",e.jsx("a",{href:"https://trace.playwright.dev",children:"trace.playwright.dev"})]}),e.jsx("div",{children:"3. Drop the trace from the download shelf into the page"})]})]}),!l&&!T&&!w&&(!c.length||y)&&e.jsxs("div",{className:"drop-target",children:[e.jsx("div",{className:"processing-error",children:y}),e.jsx("div",{className:"title",children:"Drop Playwright Trace to load"}),e.jsx("div",{children:"or"}),e.jsx("button",{onClick:()=>{const r=document.createElement("input");r.type="file",r.multiple=!0,r.click(),r.addEventListener("change",o=>W(o))},children:"Select file(s)"}),e.jsx("div",{style:{maxWidth:400},children:"Playwright Trace Viewer is a Progressive Web App, it does not send your trace anywhere, it opens it locally."})]}),l&&!c.length&&e.jsx("div",{className:"drop-target",children:e.jsx("div",{className:"title",children:"Select test to see the trace"})}),T&&e.jsx("div",{className:"drop-target",onDragLeave:()=>{h(!1)},onDrop:r=>P(r),children:e.jsx("div",{className:"title",children:"Release to analyse the Playwright Trace"})})]})},N=new b([]);(async()=>{if(C(),window.location.protocol!=="file:"){if(window.location.href.includes("isUnderTest=true")&&await new Promise(l=>setTimeout(l,1e3)),!navigator.serviceWorker)throw new Error(`Service workers are not supported.
Make sure to serve the Trace Viewer (${window.location}) via HTTPS or localhost.`);navigator.serviceWorker.register("sw.bundle.js"),navigator.serviceWorker.controller||await new Promise(l=>{navigator.serviceWorker.oncontrollerchange=()=>l()}),setInterval(function(){fetch("ping")},1e4)}F.render(e.jsx(O,{}),document.querySelector("#root"))})();
