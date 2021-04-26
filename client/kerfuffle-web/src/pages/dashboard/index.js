/*
 * Copyright (c) 2021 @nokusukun.
 * This file is part of Kerfuffle which is released under Apache.
 * See file LICENSE or go to https://github.com/nokusukun/kerfuffle/blob/master/LICENSE for full license details.
 */

import {Link, Route, Switch} from "react-router-dom";

import {
  DummyKerfuffleApplication,
  DummyKerfuffleApplication_1,
  DummyKerfuffleApplication_2
} from "../../dummy_data/dummyKerfuffleApplication";
import './dashboard.scss';
import {Button} from "primereact/button";
import {useEffect, useRef, useState} from "react";
import {Dialog} from "primereact/dialog";
import {InputText} from "primereact/inputtext";
import {deployApplication, getApplications} from "../../api/kerfuffle";
import {Messages} from "primereact/messages";
import {Application} from "../application";
import {Card} from "primereact/card";
import {Message} from "primereact/message";
import {Tag} from "primereact/tag";

export const Dashboard = () => {
  return (
    <>
      <div className={"header"}>
        <span className={"logotype"}>\kerfuffle\</span>
      </div>
      <div className="p-d-flex p-jc-center">
        <Switch>
          <Route path={"/application/:appId"}>
            <Application/>
          </Route>
          <Route path={"/"}>
            <AppListView/>
          </Route>
        </Switch>
      </div>
    </>
  )
}

const AppListView = () => {
  const [addRepo, setAddRepo] = useState(false);
  const [applications, setApplications] = useState([]);
  const [error, setError] = useState();
  const [empty, setEmpty] = useState(false);
  const [install, setInstall] = useState({
    repository: "",
    branch: "master",
    boostrap: ".kerfuffle"
  });
  const [index, setIndex] = useState()

  const [installing, setInstalling] = useState(false);
  const messageRef = useRef(null);

  useEffect(() => {
    getApplications().then(v => {
      if (v) {
        setApplications(v)
        setEmpty(false)
      } else {
        setEmpty(true)
      }
    }).catch(err => setError(err))
  }, [index])


  const deploy = () => {
    setInstalling(true);
    deployApplication(install).then(r => {
      setAddRepo(false);
      setIndex(Math.random());
    }).catch(e => {
      console.debug(e.response);
      messageRef.current.show(
        { severity: 'error', sticky: true, content: (
            <div className={"error-message"}>
              <h4>Failed to deploy</h4>
              <code>
                {e.response.data.error}
              </code>
            </div>
          ) }
      )
    }).finally(() => {
      setInstalling(false);
    })
  }

  const renderFooter = (closeFunc) => {
    return (
      <div>
        <Button label="Cancel" icon="pi pi-times" onClick={closeFunc} className="p-button-text"/>
        <Button label={installing ? "Installing..." : "Add Repository"} icon="pi pi-check" onClick={deploy} autoFocus disabled={installing}/>
      </div>
    );
  }

  return (
    <>
      <div className={"application-view"}>
        <div className={"actions"}>
          <Button label="Add Repository" className={"p-button-outlined p-button-help"}
                  onClick={() => setAddRepo(true)}/>
        </div>
        {empty &&
          <div className={"responsive p-field p-fluid"}>
            <Message severity="info" text="There's no applications. Click on Add Repository to add one."/>
          </div>
        }
        {applications.map(a => <AppViewItem application={a}/>)}
      </div>

      <Dialog header="Add New Repository" visible={addRepo} className={"responsive"}
              footer={renderFooter(() => setAddRepo(false))} onHide={() => setAddRepo(false)}>
        <div className="p-field p-fluid">
          <label htmlFor="repoURL" className="p-d-block">Repository URL <code>*required</code></label>
          <InputText value={install.repository} onInput={(e) => setInstall({...install, repository: e.target.value})} id="repoURL" aria-describedby="username1-help" className="p-d-block"/>
          <small id="repoURL-help" className="p-d-block">URL to clone</small>
        </div>
        <div className="p-field p-fluid">
          <label htmlFor="branch" className="p-d-block">Branch</label>
          <InputText value={install.branch} onInput={(e) => setInstall({...install, branch: e.target.value})} id="branch" aria-describedby="username1-help" className="p-d-block"/>
          <small id="branch-help" className="p-d-block">Branch to checkout after cloning, leave empty to
            use <code>master</code></small>
        </div>
        <div className="p-field p-fluid">
          <label htmlFor="bootstrap" className="p-d-block">Kerfuffle Bootstrap File</label>
          <InputText value={install.boostrap} onInput={(e) => setInstall({...install, bootstrap: e.target.value})} id="bootstrap" aria-describedby="username1-help" className="p-d-block"/>
          <small id="bootstrap-help" className="p-d-block">Bootstrap file to use, leave empty to
            use <code>.kerfuffle</code></small>
        </div>

        <div className="p-field p-fluid">
          <Messages ref={messageRef}/>
        </div>

      </Dialog>
    </>
  )
}

const AppViewItem = ({application: app}) => {
  const hasError = app.status !== "running"

  return (
    <Link to={`/application/${app.id}`}>
      <div className={"item"}>
        <div className={"left-info"}>
          <i className="pi pi-clone icon"/>
          {
            app.status === "running" && <Tag className="p-mr-2" value="Running"/>
          }
          {
            app.status !== "running" && <Tag className="p-mr-2" value={app.status}/>
          }
          <span>{app.meta.name || app.id}</span>
          {hasError && (
            <i className="pi pi-exclamation-circle icon-error"/>
          )}
        </div>
        <div className={"right-info"}>
          <i className="pi pi-link icon-1"/>
          <span>{app.install_configuration.repository} •&nbsp;</span>
          <span>{app.install_configuration.branch}&nbsp;</span>
        </div>
      </div>
    </Link>

  )
}