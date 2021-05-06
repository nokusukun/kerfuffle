import {CodeBlock, github} from "react-code-blocks";

export const JSONBlock = ({object}) => {
  return (
    <div style={{fontFamily: "'IBM Plex Mono'", fontSize: "0.8em"}}>
      <CodeBlock
        text={JSON.stringify(object, null, 2)}
        language={"json"}
        showLineNumbers={false}
        theme={github}
      />
    </div>
  );
}