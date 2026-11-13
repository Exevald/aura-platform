import {useEffect,useState} from 'react'
import {api,type Notification} from '../model/api'
import {NotificationsView} from '../views/ProductViews'
export function NotificationsPresenter(){const[items,setItems]=useState<Notification[]>([]),[prefs,setPrefs]=useState({inApp:true,email:true}),[error,setError]=useState('');useEffect(()=>{Promise.all([api.notifications(),api.notificationPreferences()]).then(([n,p])=>{setItems(n);setPrefs(p)}).catch(e=>setError(e instanceof Error?e.message:'Ошибка загрузки'))},[]);return <NotificationsView items={items} preferences={prefs} error={error} onRead={async id=>{await api.markNotificationRead(id);setItems(x=>x.map(v=>v.id===id?{...v,read:true}:v))}} onPreferences={async value=>setPrefs(await api.saveNotificationPreferences(value))}/>}
