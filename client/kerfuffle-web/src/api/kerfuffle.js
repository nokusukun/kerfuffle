/*
 * Copyright (c) 2021 @nokusukun.
 * This file is part of Kerfuffle which is released under Apache.
 * See file LICENSE or go to https://github.com/nokusukun/kerfuffle/blob/master/LICENSE for full license details.
 */

import axios from "axios";

const API_URL = window.location.host === "localhost:3000" ? "http://localhost:8080" : window.location.origin
const v1 = APIGenerator(API_URL).api.v1

function _sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

export async function getApplications() {
  const response = await axios.get(v1.application)
  return response.data
}

export async function getApplication(id) {
  const response = await axios.get(v1.$application(id))
  return response.data
}

export async function deployApplication(installData) {
  console.log("Posting", v1.application, installData)
  const response = await axios.post(v1.application, installData)
  return response.data
}

export async function holdApplication(id) {
  const response = await axios.patch(v1.$application(id).hold)
  return response.data
}

export async function reloadApplication(id) {
  const response = await axios.patch(v1.$application(id).reload)
  return response.data
}

export async function shutdownApplication(id) {
  const response = await axios.patch(v1.$application(id).shutdown)
  return response.data
}

export async function startupApplication(id) {
  const response = await axios.patch(v1.$application(id).startup)
  return response.data
}

export async function deleteApplication(id) {
  const response = await axios.delete(v1.$application(id))
  return response.data
}

export async function getLog(id, provision, log_type, from) {
  const response = await axios.get(v1.$application(id).$provision(provision).$output(log_type).toString() + `?from=${from}`)
  return response.data
}


function APIGenerator(...urls) {
  const handler = {
    get: function (object, prop) {
      if (!Object.keys(object).includes(prop)) {
        if (prop[0] === "$") {
          return (parameter) => {
            return APIGenerator(...object.paths, prop.substr(1), parameter)
          }
        }
        return APIGenerator(...object.paths, prop)
      }
      return object[prop]
    }
  }
  const paths = {
    paths: urls,
    toString: function () {
      return this.paths.join("/")
    },
    param: function (p) {
      this.paths.push(p)
      return this
    }
  }

  return new Proxy(paths, handler)
}
