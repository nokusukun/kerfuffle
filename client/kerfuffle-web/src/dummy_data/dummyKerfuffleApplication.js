/*
 * Copyright (c) 2021 @nokusukun.
 * This file is part of Kerfuffle which is released under Apache.
 * See file LICENSE or go to https://github.com/nokusukun/kerfuffle/blob/master/LICENSE for full license details.
 */

/*
 * Copyright (c) 2021 @nokusukun.
 * This file is part of Kerfuffle which is released under Apache.
 * See file LICENSE or go to https://github.com/nokusukun/kerfuffle/blob/master/LICENSE for full license details.
 */

export const DummyKerfuffleApplication_2 = {
  "application": {
    "id": "github-com-nokusukun-sample-express@develop",
    "install_configuration": {
      "repository": "https://github.com/nokusukun/sample-express",
      "branch": "develop",
      "bootstrap": "develop.kerfuffle"
    },
    "meta": {
      "name": "[develop] Sample Express Application"
    }
  },
  "cfs": {
    "backend": {
      "Host": [
        "api.example.com",
        "api.www.example.com"
      ],
      "Zone": "example.com",
      "Proxied": true
    },
    "frontend": {
      "Host": [
        "www.example.com"
      ],
      "Zone": "example.com",
      "Proxied": true
    }
  },
  "processes": {
    "backend": {
      "alive": true,
      "status": "C:\\Users\\defyh\\AppData\\Roaming\\npm\\yarn.cmd run run"
    },
    "frontend": {
      "alive": true,
      "status": "C:\\Users\\defyh\\AppData\\Roaming\\npm\\yarn.cmd run run"
    }
  },
  "provisions": {
    "backend": {
      "id": "backend",
      "health_endpoint": "/",
      "event_url": "https://postb.in/1602269478542-1194403597619",
      "run": [
        [
          "yarn"
        ],
        [
          "yarn",
          "run",
          "run"
        ]
      ],
      "environment_variables": [
        "MESSAGE=BACKEND"
      ]
    },
    "frontend": {
      "id": "frontend",
      "health_endpoint": "/",
      "event_url": "https://postb.in/1602269478542-1194403597619",
      "run": [
        [
          "yarn"
        ],
        [
          "yarn",
          "run",
          "run"
        ]
      ],
      "environment_variables": [
        "MESSAGE=FRONTEND"
      ]
    }
  },
  "proxies": {
    "backend": {
      "host": [
        "api.example.com",
        "api.www.example.com"
      ],
      "bind_port": "15615"
    },
    "frontend": {
      "host": [
        "www.example.com"
      ],
      "bind_port": "15614"
    }
  }
}

export const DummyKerfuffleApplication_1 = {
  "application": {
    "id": "github-com-nokusukun-sample-express@master",
    "install_configuration": {
      "repository": "https://github.com/nokusukun/sample-express",
      "branch": "master",
      "bootstrap": ".kerfuffle"
    },
    "meta": {
      "name": "[production] Sample Express Application"
    }
  },
  "cfs": {
    "backend": {
      "Host": [
        "api.example.com",
        "api.www.example.com"
      ],
      "Zone": "example.com",
      "Proxied": true
    },
    "frontend": {
      "Host": [
        "www.example.com"
      ],
      "Zone": "example.com",
      "Proxied": true
    }
  },
  "processes": {
    "backend": {
      "alive": true,
      "status": "C:\\Users\\defyh\\AppData\\Roaming\\npm\\yarn.cmd run run"
    },
    "frontend": {
      "alive": true,
      "status": "C:\\Users\\defyh\\AppData\\Roaming\\npm\\yarn.cmd run run"
    }
  },
  "provisions": {
    "backend": {
      "id": "backend",
      "health_endpoint": "/",
      "event_url": "https://postb.in/1602269478542-1194403597619",
      "run": [
        [
          "yarn"
        ],
        [
          "yarn",
          "run",
          "run"
        ]
      ],
      "environment_variables": [
        "MESSAGE=BACKEND"
      ]
    },
    "frontend": {
      "id": "frontend",
      "health_endpoint": "/",
      "event_url": "https://postb.in/1602269478542-1194403597619",
      "run": [
        [
          "yarn"
        ],
        [
          "yarn",
          "run",
          "run"
        ]
      ],
      "environment_variables": [
        "MESSAGE=FRONTEND"
      ]
    }
  },
  "proxies": {
    "backend": {
      "host": [
        "api.example.com",
        "api.www.example.com"
      ],
      "bind_port": "15615"
    },
    "frontend": {
      "host": [
        "www.example.com"
      ],
      "bind_port": "15614"
    }
  }
}


export const DummyKerfuffleApplication = {
  "application": {
    "id": "gitlab-com-kicph-proctor-platform@master",
    "install_configuration": {
      "repository": "https://gitlab.com/kicph/proctor-platform",
      "branch": "master",
      "bootstrap": ".kerfuffle"
    },
    "meta": {
      "name": "[production] Proctor Platform"
    }
  },
  "cfs": {
    "backend": {
      "Host": [
        "api.example.com",
        "api.www.example.com"
      ],
      "Zone": "example.com",
      "Proxied": true
    },
    "frontend": {
      "Host": [
        "www.example.com"
      ],
      "Zone": "example.com",
      "Proxied": true
    }
  },
  "processes": {
    "backend": {
      "alive": false,
      "status": "C:\\Users\\defyh\\AppData\\Roaming\\npm\\yarn.cmd run run"
    },
    "frontend": {
      "alive": true,
      "status": "C:\\Users\\defyh\\AppData\\Roaming\\npm\\yarn.cmd run run"
    }
  },
  "provisions": {
    "backend": {
      "id": "backend",
      "health_endpoint": "/",
      "event_url": "https://postb.in/1602269478542-1194403597619",
      "run": [
        [
          "yarn"
        ],
        [
          "yarn",
          "run",
          "run"
        ]
      ],
      "environment_variables": [
        "MESSAGE=BACKEND"
      ]
    },
    "frontend": {
      "id": "frontend",
      "health_endpoint": "/",
      "event_url": "https://postb.in/1602269478542-1194403597619",
      "run": [
        [
          "yarn"
        ],
        [
          "yarn",
          "run",
          "run"
        ]
      ],
      "environment_variables": [
        "MESSAGE=FRONTEND"
      ]
    }
  },
  "proxies": {
    "backend": {
      "host": [
        "api.example.com",
        "api.www.example.com"
      ],
      "bind_port": "15615"
    },
    "frontend": {
      "host": [
        "www.example.com"
      ],
      "bind_port": "15614"
    }
  }
}