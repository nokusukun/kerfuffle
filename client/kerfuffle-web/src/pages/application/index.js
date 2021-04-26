/*
 * Copyright (c) 2021 @nokusukun.
 * This file is part of Kerfuffle which is released under Apache.
 * See file LICENSE or go to https://github.com/nokusukun/kerfuffle/blob/master/LICENSE for full license details.
 */

import { useParams } from "react-router-dom";
import {useEffect, useRef, useState} from "react";
import {getApplication, holdApplication} from "../../api/kerfuffle";
import {Messages} from "primereact/messages";
import {BreadCrumb} from "primereact/breadcrumb";

import './application.scss';
import {TabPanel, TabView} from "primereact/tabview";
import {Panel} from "primereact/panel";
import {InputText} from "primereact/inputtext";
import {Button} from "primereact/button";
import {CodeBlock, github} from "react-code-blocks";
import {Message} from "primereact/message";
import {Divider} from "primereact/divider";
import {Toast} from "primereact/toast";

const JSONBlock = ({ object }) => {
  return (
      <div style={{fontFamily: "'IBM Plex Mono'", fontSize: "0.8em"}}>
        <CodeBlock
            text={ JSON.stringify(object, null, 2)}
            language={"json"}
            showLineNumbers={false}
            theme={github}
        />
      </div>
  );
}


const MessageContent = ({header, data}) => {
  return (
    <div style={{display: 'flex', flexDirection: 'column'}}>
      <h4>{header}</h4>
      <p>
        <code>
          {data}
        </code>
      </p>
    </div>
  )
}


const Spacer = ({height}) => {
  return (<div style={{height: height || "1em"}}/>)
}


function CloudflarePanel({app}) {

  return <div>
    <JSONBlock object={app.cfs}/>
  </div>
}


function ProxiesPanel({app}) {

  return <div>
    <JSONBlock object={app.proxies}/>
  </div>
}


function ProvisionsPanel({app}) {

  return <div>
    {Object.values(app.provisions).map(prov =>
      <>
        <Spacer height={"1em"}/>
        <Panel header={prov.id.toString().toUpperCase()}>
          <JSONBlock object={prov}/>
        </Panel>
      </>
    )}
  </div>
}

function OverviewPanel({app, reload}) {
  const toast = useRef(null);

  const holdApp = () => {
    holdApplication(app.application.id).then(v => {
      toast.current.show({severity:'info', summary:'Success', detail:'Toggled maintenance mode', life: 3000})
      reload && reload()
    }).catch(e => {
      toast.current.show({severity:'error', summary:'Failed', detail:'Failed to toggle maintenance mode', life: 3000})
    })
  }

  const lastLog = app.application.status_log && app.application.status_log[app.application.status_log.length - 1]
  const colorway = {
    booting: "warn",
    running: "info",
    failed: "error",
    crashed: "error",
    unknown: ""
  }

  return <div className="p-grid">
    <div className="p-col-8">
      <Toast ref={toast} />
      <Panel header="Information">
        <div className="p-field p-fluid">
          <label className="p-d-block"><b>ID</b></label>
          <InputText value={app.application.id} disabled className="p-d-block"/>
          <Spacer height={"0.8em"}/>
          <label className="p-d-block"><b>Repository</b></label>
          <InputText value={app.application.install_configuration.repository} disabled className="p-d-block"/>
          <Spacer height={"0.8em"}/>
          <label className="p-d-block"><b>Branch</b></label>
          <InputText value={app.application.install_configuration.branch} disabled className="p-d-block"/>
          <Spacer height={"0.8em"}/>
          <label className="p-d-block"><b>Bootstrap</b></label>
          <InputText value={app.application.install_configuration.bootstrap} disabled className="p-d-block"/>
          <Spacer height={"0.8em"}/>
          <label className="p-d-block"><b>Meta.name</b></label>
          <InputText value={app.application.meta.name} disabled className="p-d-block"/>
        </div>
      </Panel>
      <Spacer />
      <Panel header={"Provision Status"}>
        <JSONBlock object={app.processes}/>
      </Panel>

    </div>
    <div className="p-col-4">
      {
        lastLog && lastLog.flag === "running" &&
        <div className={"p-fluid"}>
          <Message severity={colorway[lastLog.flag]} text={lastLog.reason} />
          <Spacer/>
        </div>
      }
      {
        app.application.maintenance_mode &&
        <div className={"p-fluid"}>
          <Message severity="warn" text="This application is in maintenance mode" />
          <Spacer/>
        </div>
      }
      <Panel header={"Actions"}>
        <Button label="Reload" className="p-button-outlined" />
        <Button label="Shutdown" className="p-button-outlined p-button-help" />
        <Button label="Delete" className="p-button-outlined p-button-danger" />

        <Divider align="left" >
          <b>Maintenance Mode</b>
        </Divider>
        {app.application.maintenance_mode ?
            <Button label="Go Live" onClick={holdApp} className="p-button-outlined p-button-success" /> :
            <Button label="Hold Application" onClick={holdApp} className="p-button-outlined p-button-warning" />
        }
      </Panel>
      <Spacer/>
      <Panel header={"Last Commit"}>
        <div style={{whiteSpace: "pre-line", fontFamily: "'IBM Plex mono'", fontSize: "0.8em"}}>
          {app.last_commit}
        </div>
      </Panel>

    </div>
  </div>;
}

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
    getApplication(appId).then(
      app => {
        setApp(app)
        setBreadItems(
          [
            {label: app.application.id}
          ]
        )
      }
    ).catch( e => {
        messageRef.current.show({
          severity: 'error',
          sticky: true,
          content: <MessageContent header={"Failed to load"} data={e.response.data.error} />})
      }
    ).finally(() => setLoading(false))
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
        <i className="pi pi-spin pi-spinner" style={{'fontSize': '2em'}}/>
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
