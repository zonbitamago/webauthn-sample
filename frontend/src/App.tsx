import { useState } from "react";
import axios from "axios";

const API_ROOT_PATH = "http://localhost/api";

function App() {
  const [name, setName] = useState("");

  // Base64 to ArrayBuffer
  const bufferDecode = (value: string) => {
    return Uint8Array.from(atob(value), (c) => c.charCodeAt(0));
  };

  // ArrayBuffer to URLBase64
  function bufferEncode(value: any) {
    // @ts-ignore
    return btoa(String.fromCharCode.apply(null, new Uint8Array(value)))
      .replace(/\+/g, "-")
      .replace(/\//g, "_")
      .replace(/=/g, "");
  }

  const registrationHandler = async () => {
    const result = await axios.get(`${API_ROOT_PATH}/register/begin/${name}`, {
      withCredentials: true,
    });
    const data = result.data;

    data.publicKey.challenge = bufferDecode(data.publicKey.challenge);
    data.publicKey.user.id = bufferDecode(data.publicKey.user.id);

    if (data.publicKey.excludeCredentials) {
      for (let i = 0; i < data.publicKey.excludeCredentials.length; i++) {
        data.publicKey.excludeCredentials[i].id = bufferDecode(
          data.publicKey.excludeCredentials[i].id
        );
      }
    }

    const credential = await navigator.credentials.create({
      publicKey: data.publicKey,
    });

    // @ts-ignore
    const attestationObject = credential.response.attestationObject;
    // @ts-ignore
    const clientDataJSON = credential.response.clientDataJSON;
    // @ts-ignore
    const rawId = credential.rawId;

    try {
      const finishData = await axios.post(
        `${API_ROOT_PATH}/register/finish/${name}`,
        {
          id: credential?.id,
          rawId: bufferEncode(rawId),
          type: credential?.type,
          response: {
            attestationObject: bufferEncode(attestationObject),
            clientDataJSON: bufferEncode(clientDataJSON),
          },
        },
        { withCredentials: true }
      );
      alert(`successfully registered ${name}!`);
    } catch (error) {
      console.log(error);
      alert(`failed to register ${name}`);
    }
  };

  return (
    <div>
      <div>
        name:
        <input
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
        />
        <br />
        <button onClick={registrationHandler}>Registration</button>
      </div>
    </div>
  );
}

export default App;
