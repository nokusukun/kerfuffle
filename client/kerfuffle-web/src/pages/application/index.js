/*
 * Copyright (c) 2021 @nokusukun.
 * This file is part of Kerfuffle which is released under Apache.
 * See file LICENSE or go to https://github.com/nokusukun/kerfuffle/blob/master/LICENSE for full license details.
 */

import {useParams} from "react-router-dom";
import {useEffect, useRef, useState} from "react";
import {getApplication} from "../../api/kerfuffle";
import {Messages} from "primereact/messages";
import {BreadCrumb} from "primereact/breadcrumb";

import './application.scss';
import {TabPanel, TabView} from "primereact/tabview";
import {ProgressBar} from "primereact/progressbar";
import {CloudflarePanel, MessageContent, OverviewPanel, ProvisionsPanel, ProxiesPanel} from "./panels";


export const Application = () => {
  let { appId } = useParams();
  const [app, setApp] = useState();
  const [loading, setLoading] = useState(false);
  const [nonce, setNonce] = useState();
  const messageRef = useRef(null);

  const [breadItems, setBreadItems] = useState([
    {label: 'Loading...'}
  ]);
  const home = { icon: 'pi pi-cog', url: '/console' }

  useEffect(() => {
    setLoading(true)
    const interv = setInterval(function pulldata() {
      getApplication(appId).then(
        app => {
          setApp(app)
          setBreadItems([{label: app.application.id}])
        }
      ).catch( e => {
        if (!messageRef.current) {
          return
        }
        messageRef.current.show({
          severity: 'error',
          sticky: true,
          content: <MessageContent header={"Failed to load"} data={e.response.data.error} />})
      }).finally(() => setLoading(false))
      return pulldata
    }(), 5000)

    return () => clearInterval(interv)
  }, [appId, nonce])

  const reload = () => {
    setNonce(Math.random())
  }

  return (
    <div className={"app-view-panel"}>
      <BreadCrumb model={breadItems} home={home}/>
      <div className="p-field p-fluid">
        <Messages ref={messageRef}/>
      </div>
      { loading &&
      <div className={"p-field pi-fluid fl-center"}>
        <ProgressBar mode="indeterminate" />
      </div>
      }
      {
        app &&
        <TabView>
          <TabPanel header="Overview">
            <OverviewPanel app={app} reload={reload}/>
          </TabPanel>
          <TabPanel header="Provisions">
            <ProvisionsPanel app={app}/>
          </TabPanel>
          <TabPanel header="Reverse Proxies">
            <ProxiesPanel app={app}/>
          </TabPanel>
          <TabPanel header="Cloudflare">
            <CloudflarePanel app={app}/>
          </TabPanel>
        </TabView>
      }

    </div>
  )
}
