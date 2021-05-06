import {useEffect, useRef, useState} from "react";
import {
  deleteApplication,
  getLog,
  holdApplication,
  reloadApplication,
  shutdownApplication,
  startupApplication
} from "../../api/kerfuffle";
import {Dialog} from "primereact/dialog";
import {ProgressBar} from "primereact/progressbar";
import {CodeBlock, dracula} from "react-code-blocks";
import {Toast} from "primereact/toast";
import {Panel} from "primereact/panel";
import {Spacer} from "./spacer";
import {DataTable} from "primereact/datatable";
import {Column} from "primereact/column";
import {Tag} from "primereact/tag";
import {Timeline} from "primereact/timeline";
import {Message} from "primereact/message";
import {Button} from "primereact/button";
import {Divider} from "primereact/divider";
import {Menu} from "primereact/menu";
import {JSONBlock} from "./JSONBlock";

export const MessageContent = ({header, data}) => {
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

export function CloudflarePanel({app}) {

  return <div>
    <JSONBlock object={app.cfs}/>
  </div>
}

export function ProxiesPanel({app}) {

  return <div>
    <JSONBlock object={app.proxies}/>
  </div>
}

export function ProvisionsPanel({app}) {

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

function LogDisplayModal({app_id, provision, log_type, close}) {
  const [log, setLog] = useState("")
  const [from, setFrom] = useState(0)
  const interval = useRef(null)

  console.log(app_id, provision, log_type, from)

  useEffect(() => {
    console.log(app_id, provision, log_type, from)

    interval.current = setInterval(function _self() {
      getLog(app_id, provision, log_type, from).then(result => {
        setFrom(result.next)
        setLog(log + result.content)
      })
      return _self
    }(), 1000)

    return () => {
      setLog("")
      setFrom(0)
      clearInterval(interval.current)
    }
  }, [app_id, provision, log_type])


  return <Dialog
    header={`'${log_type}' Output View: ${provision}`}
    visible={true}
    onHide={close}
    style={{whiteSpace: "pre-line", fontFamily: "'IBM Plex mono'", fontSize: "0.8em", width: "50vw"}}>
    {!log && <ProgressBar mode="indeterminate"/>}
    <CodeBlock
      text={log}
      showLineNumbers={false}
      theme={dracula}
    />
  </Dialog>
}

export function OverviewPanel({app, reload}) {
  const toast = useRef(null);
  const [logModal, setLogModal] = useState(null);
  const [loading, setLoading] = useState(false);

  const holdApp = () => {
    holdApplication(app.application.id).then(v => {
      toast.current.show({severity: 'info', summary: 'Success', detail: 'Toggled maintenance mode', life: 3000})
      reload && reload()
    }).catch(e => {
      toast.current.show({
        severity: 'error',
        summary: 'Failed',
        detail: 'Failed to toggle maintenance mode',
        life: 3000
      })
    })
  }

  const reloadApp = () => {
    setLoading(true)
    reloadApplication(app.application.id).then(v => {
      toast.current.show({severity: 'info', summary: 'Success', detail: 'Reloading Application', life: 3000})
      reload && reload()
    }).catch(e => {
      toast.current.show({
        severity: 'error',
        summary: 'Failed',
        detail: `Action Failed: ${e.response.data.error}`,
        life: 3000
      })
    }).finally(() => setLoading(false))
  }

  const shutdownApp = () => {
    shutdownApplication(app.application.id).then(v => {
      toast.current.show({severity: 'info', summary: 'Success', detail: 'Shutting down Application', life: 3000})
      reload && reload()
    }).catch(e => {
      toast.current.show({
        severity: 'error',
        summary: 'Failed',
        detail: `Action Failed: ${e.response.data.error}`,
        life: 3000
      })
    })
  }

  const startApp = () => {
    startupApplication(app.application.id).then(v => {
      toast.current.show({severity: 'info', summary: 'Success', detail: 'Starting Application', life: 3000})
      reload && reload()
    }).catch(e => {
      toast.current.show({
        severity: 'error',
        summary: 'Failed',
        detail: `Action Failed: ${e.response.data.error}`,
        life: 3000
      })
    })
  }

  const deleteApp = () => {
    deleteApplication(app.application.id).then(v => {
      toast.current.show({severity: 'info', summary: 'Success', detail: 'Deleting Application', life: 3000})
      reload && reload()
    }).catch(e => {
      toast.current.show({
        severity: 'error',
        summary: 'Failed',
        detail: `Action Failed: ${e.response.data.error}`,
        life: 3000
      })
    })
  }

  const lastLog = app.application.status_log && app.application.status_log[0]
  const colorway = {
    booting: "warn",
    running: "info",
    failed: "error",
    crashed: "error",
    shutdown: "warn",
    unknown: "",
  }

  const menu_log_stdout = useRef(null);
  const menu_log_stderr = useRef(null);

  const closeModal = () => {
    setLogModal(null)
  }

  const processStatus = Object.entries(app.processes).map(v => {
    const [key, value] = v
    return {...value, id: key}
  })

  const renderLogMenuItems = (logType) => Object.values(app.provisions).map(prov => ({
    label: prov.id,
    command: () => {
      setLogModal(<LogDisplayModal close={closeModal} app_id={app.application.id} provision={prov.id}
                                   log_type={logType}/>)
    }
  }))

  return <div className="p-grid">
    <div className="p-col-8">
      <Toast ref={toast}/>
      <Panel header="Information">
        <div className={"p-grid"}>
          <div className={'p-col-6'}>
            <dl>
              <dt>ID</dt>
              <dd>{app.application.id}</dd>

              <dt>Repository</dt>
              <dd><a target={'_blank'} href={app.application.install_configuration.repository}>
                {app.application.install_configuration.repository}</a>
              </dd>

              <dt>Branch</dt>
              <dd>{app.application.install_configuration.branch}</dd>
            </dl>
          </div>
          <div className={'p-col-6'}>
            <dl>
              <dt>Bootstrap</dt>
              <dd>{app.application.install_configuration.bootstrap}</dd>

              <dt>Meta.name</dt>
              <dd>{app.application.meta.name}</dd>
              <dt>Physical Location</dt>
              <dd>{app.application.root_path}</dd>
            </dl>
          </div>
        </div>
      </Panel>
      <Spacer/>
      <Panel header={"Process Status"}>
        <DataTable value={processStatus} className="p-datatable-sm">
          <Column field="id" style={{width: '15%', fontWeight: 'bold'}} header="Provision"/>
          <Column header="State"
                  style={{width: '10%'}}
                  body={value => value.alive ? <Tag className="p-mr-2" value="Alive" severity={"success"}/> :
                    <Tag className="p-mr-2" value="Dead" severity="danger"/>}
          />
          <Column header="Status" body={data => <code>{data.status}</code>}/>
        </DataTable>
      </Panel>
      <Spacer/>
      <Panel header={"Events"}>
        <Timeline value={app.application.status_log} opposite={(item) => item.reason}
                  content={(item) => <small className="p-text-secondary">{new Date(item.at).toGMTString()}</small>}/>
      </Panel>
    </div>
    <div className="p-col-4">
      {
        lastLog &&
        <div className={"p-fluid"}>
          <Message severity={colorway[lastLog.flag]} text={lastLog.reason}/>
          <Spacer/>
        </div>
      }
      {
        app.application.maintenance_mode &&
        <div className={"p-fluid"}>
          <Message severity="warn" text="This application is in maintenance mode"/>
          <Spacer/>
        </div>
      }
      <Panel header={"Actions"}>
        {loading && <ProgressBar mode="indeterminate"/>}
        <Button label="Reload" className="p-button-outlined" onClick={reloadApp}/>
        {
          lastLog && lastLog.flag === "running" &&
          <Button label="Shutdown" className="p-button-outlined p-button-help" onClick={shutdownApp}/>
        }
        {
          lastLog && lastLog.flag === "shutdown" &&
          <Button label="Start" className="p-button-outlined p-button-help" onClick={startApp}/>
        }
        <Button label="Delete" className="p-button-outlined p-button-danger" onClick={deleteApp}/>

        <Divider align="left">
          <b>Maintenance Mode</b>
        </Divider>
        {app.application.maintenance_mode ?
          <Button label="Go Live" onClick={holdApp} className="p-button-outlined p-button-success"/> :
          <Button label="Hold Application" onClick={holdApp} className="p-button-outlined p-button-warning"/>
        }

        <Divider align="left">
          <b>Logs</b>
        </Divider>
        {logModal && logModal}
        <Menu model={renderLogMenuItems("log")} popup ref={menu_log_stdout} id="popup_menu"/>
        <Button label="Stdout" icon="pi pi-bars" onClick={(event) => menu_log_stdout.current.toggle(event)}
                aria-controls="popup_menu" aria-haspopup/>

        <Menu model={renderLogMenuItems("err")} popup ref={menu_log_stderr} id="popup_menu"/>
        <Button label="Stderr" icon="pi pi-bars" onClick={(event) => menu_log_stderr.current.toggle(event)}
                aria-controls="popup_menu" aria-haspopup/>
      </Panel>

      <Spacer/>
      <Panel header={"Last Commit"}>
        <div style={{whiteSpace: "pre-line", fontFamily: "'IBM Plex mono'", fontSize: "0.8em"}}>
          {app.last_commit}
        </div>
      </Panel>
      <Spacer/>

    </div>
  </div>;
}