import{d as I,y as N,u as A,a as B,N as C,V as x,k as r,aB as V,o as w,b as m,e,a4 as g,aP as L,t as M,f as d,m as y,_ as R,g as U,w as Y,l as P,v as F,p as $,j as q}from"./index.d56ab597.js";import{u as z,_ as E}from"./Modal.vue_vue_type_script_setup_true_lang.3eabc9b1.js";import{s as S}from"./index.985d3c1a.js";import{_ as T}from"./_plugin-vue_export-helper.cdc0426e.js";const l=n=>($("data-v-a0290a9d"),n=n(),q(),n),G={class:"container flex flex-col w-full"},H=l(()=>e("h1",{class:"text-3xl mb-3"},"Download the Files",-1)),J={class:"file-list w-full bg-gray-800 flex flex-col justify-center items-start overflow-y-auto"},K={class:"flex flex-col justify-start items-start"},O=l(()=>e("span",{class:"text-green-500 font-bold"},">",-1)),Q={class:"ml-1"},W={class:"flex flex-row justify-center items-center w-full mt-3"},X=["disabled","onClick"],Z=["onSubmit"],ee={class:"form-control textbox"},se=l(()=>e("span",null,"Download Page Password",-1)),te={class:"form-control textbox"},oe=l(()=>e("span",null,"Your Account Password",-1)),ae=l(()=>e("div",{class:"flex flex-row justify-end items-center w-full"},[e("button",{type:"submit",class:"block btn btn-orange"},"Show File")],-1)),le=I({__name:"FileDownload",setup(n){const h=N(),v=A(),u=z(),f=B();f.isAuthenticated||f.validateAuth();const b=C({code:h.params.code,files:[]}),_=x(b,"code"),o=x(b,"files"),p=r(!1),k=r(!1);V(async()=>{const s=await u.viewFiles(_.value);if(s||v.push({name:"error"}),s.isNeedLogin&&!f.isAuthenticated){v.push({name:"login-account",query:{url:h.fullPath}});return}if(!s.isAllowed){S("You're not allowed to open this file");return}p.value=s.isProtected,k.value=s.isNeedLogin,s.isProtected||(o.value=s.files)});const i=r(),c=r(),D=async()=>{const{ok:s,message:a,files:t}=await u.viewProtectedFiles(_.value,{pagePassword:i.value,accountPassword:c.value});if(!s){S(a);return}o.value=t,p.value=!1},j=async()=>{await u.downloadFiles(_.value,{pagePassword:i.value,accountPassword:c.value})};return(s,a)=>(w(),m(g,null,[e("div",G,[H,e("div",J,[e("ul",K,[(w(!0),m(g,null,L(d(o),t=>(w(),m("li",{class:"mb-2",key:t.fileId},[O,e("span",Q,M(t.fileName),1)]))),128))])]),e("div",W,[e("button",{disabled:d(o).length<=0,onClick:y(j,["prevent"]),class:R(["block btn",{"btn-orange":d(o).length>0,"btn-disable":d(o).length<=0}])},"Download",10,X)])]),U(E,{show:p.value,"no-close-button":!0},{default:Y(()=>[e("form",{onSubmit:y(D,["prevent"]),class:"flex flex-col justify-start items-start mt-3"},[e("label",ee,[se,P(e("input",{type:"password","onUpdate:modelValue":a[0]||(a[0]=t=>i.value=t)},null,512),[[F,i.value]])]),e("label",te,[oe,P(e("input",{type:"password","onUpdate:modelValue":a[1]||(a[1]=t=>c.value=t)},null,512),[[F,c.value]])]),ae],40,Z)]),_:1},8,["show"])],64))}});const de=T(le,[["__scopeId","data-v-a0290a9d"]]);export{de as default};