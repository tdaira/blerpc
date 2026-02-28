import * as $protobuf from 'protobufjs';
import Long = require('long');
/** Namespace blerpc. */
export namespace blerpc {
  /** Properties of an EchoRequest. */
  interface IEchoRequest {
    /** EchoRequest message */
    message?: string | null;
  }

  /** Represents an EchoRequest. */
  class EchoRequest implements IEchoRequest {
    /**
     * Constructs a new EchoRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: blerpc.IEchoRequest);

    /** EchoRequest message. */
    public message: string;

    /**
     * Creates a new EchoRequest instance using the specified properties.
     * @param [properties] Properties to set
     * @returns EchoRequest instance
     */
    public static create(properties?: blerpc.IEchoRequest): blerpc.EchoRequest;

    /**
     * Encodes the specified EchoRequest message. Does not implicitly {@link blerpc.EchoRequest.verify|verify} messages.
     * @param message EchoRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: blerpc.IEchoRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Encodes the specified EchoRequest message, length delimited. Does not implicitly {@link blerpc.EchoRequest.verify|verify} messages.
     * @param message EchoRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encodeDelimited(
      message: blerpc.IEchoRequest,
      writer?: $protobuf.Writer,
    ): $protobuf.Writer;

    /**
     * Decodes an EchoRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns EchoRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(
      reader: $protobuf.Reader | Uint8Array,
      length?: number,
    ): blerpc.EchoRequest;

    /**
     * Decodes an EchoRequest message from the specified reader or buffer, length delimited.
     * @param reader Reader or buffer to decode from
     * @returns EchoRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decodeDelimited(reader: $protobuf.Reader | Uint8Array): blerpc.EchoRequest;

    /**
     * Verifies an EchoRequest message.
     * @param message Plain object to verify
     * @returns `null` if valid, otherwise the reason why it is not
     */
    public static verify(message: { [k: string]: any }): string | null;

    /**
     * Creates an EchoRequest message from a plain object. Also converts values to their respective internal types.
     * @param object Plain object
     * @returns EchoRequest
     */
    public static fromObject(object: { [k: string]: any }): blerpc.EchoRequest;

    /**
     * Creates a plain object from an EchoRequest message. Also converts values to other types if specified.
     * @param message EchoRequest
     * @param [options] Conversion options
     * @returns Plain object
     */
    public static toObject(
      message: blerpc.EchoRequest,
      options?: $protobuf.IConversionOptions,
    ): { [k: string]: any };

    /**
     * Converts this EchoRequest to JSON.
     * @returns JSON object
     */
    public toJSON(): { [k: string]: any };

    /**
     * Gets the default type url for EchoRequest
     * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
     * @returns The default type url
     */
    public static getTypeUrl(typeUrlPrefix?: string): string;
  }

  /** Properties of an EchoResponse. */
  interface IEchoResponse {
    /** EchoResponse message */
    message?: string | null;
  }

  /** Represents an EchoResponse. */
  class EchoResponse implements IEchoResponse {
    /**
     * Constructs a new EchoResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: blerpc.IEchoResponse);

    /** EchoResponse message. */
    public message: string;

    /**
     * Creates a new EchoResponse instance using the specified properties.
     * @param [properties] Properties to set
     * @returns EchoResponse instance
     */
    public static create(properties?: blerpc.IEchoResponse): blerpc.EchoResponse;

    /**
     * Encodes the specified EchoResponse message. Does not implicitly {@link blerpc.EchoResponse.verify|verify} messages.
     * @param message EchoResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(
      message: blerpc.IEchoResponse,
      writer?: $protobuf.Writer,
    ): $protobuf.Writer;

    /**
     * Encodes the specified EchoResponse message, length delimited. Does not implicitly {@link blerpc.EchoResponse.verify|verify} messages.
     * @param message EchoResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encodeDelimited(
      message: blerpc.IEchoResponse,
      writer?: $protobuf.Writer,
    ): $protobuf.Writer;

    /**
     * Decodes an EchoResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns EchoResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(
      reader: $protobuf.Reader | Uint8Array,
      length?: number,
    ): blerpc.EchoResponse;

    /**
     * Decodes an EchoResponse message from the specified reader or buffer, length delimited.
     * @param reader Reader or buffer to decode from
     * @returns EchoResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decodeDelimited(reader: $protobuf.Reader | Uint8Array): blerpc.EchoResponse;

    /**
     * Verifies an EchoResponse message.
     * @param message Plain object to verify
     * @returns `null` if valid, otherwise the reason why it is not
     */
    public static verify(message: { [k: string]: any }): string | null;

    /**
     * Creates an EchoResponse message from a plain object. Also converts values to their respective internal types.
     * @param object Plain object
     * @returns EchoResponse
     */
    public static fromObject(object: { [k: string]: any }): blerpc.EchoResponse;

    /**
     * Creates a plain object from an EchoResponse message. Also converts values to other types if specified.
     * @param message EchoResponse
     * @param [options] Conversion options
     * @returns Plain object
     */
    public static toObject(
      message: blerpc.EchoResponse,
      options?: $protobuf.IConversionOptions,
    ): { [k: string]: any };

    /**
     * Converts this EchoResponse to JSON.
     * @returns JSON object
     */
    public toJSON(): { [k: string]: any };

    /**
     * Gets the default type url for EchoResponse
     * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
     * @returns The default type url
     */
    public static getTypeUrl(typeUrlPrefix?: string): string;
  }

  /** Properties of a FlashReadRequest. */
  interface IFlashReadRequest {
    /** FlashReadRequest address */
    address?: number | null;

    /** FlashReadRequest length */
    length?: number | null;
  }

  /** Represents a FlashReadRequest. */
  class FlashReadRequest implements IFlashReadRequest {
    /**
     * Constructs a new FlashReadRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: blerpc.IFlashReadRequest);

    /** FlashReadRequest address. */
    public address: number;

    /** FlashReadRequest length. */
    public length: number;

    /**
     * Creates a new FlashReadRequest instance using the specified properties.
     * @param [properties] Properties to set
     * @returns FlashReadRequest instance
     */
    public static create(properties?: blerpc.IFlashReadRequest): blerpc.FlashReadRequest;

    /**
     * Encodes the specified FlashReadRequest message. Does not implicitly {@link blerpc.FlashReadRequest.verify|verify} messages.
     * @param message FlashReadRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(
      message: blerpc.IFlashReadRequest,
      writer?: $protobuf.Writer,
    ): $protobuf.Writer;

    /**
     * Encodes the specified FlashReadRequest message, length delimited. Does not implicitly {@link blerpc.FlashReadRequest.verify|verify} messages.
     * @param message FlashReadRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encodeDelimited(
      message: blerpc.IFlashReadRequest,
      writer?: $protobuf.Writer,
    ): $protobuf.Writer;

    /**
     * Decodes a FlashReadRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns FlashReadRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(
      reader: $protobuf.Reader | Uint8Array,
      length?: number,
    ): blerpc.FlashReadRequest;

    /**
     * Decodes a FlashReadRequest message from the specified reader or buffer, length delimited.
     * @param reader Reader or buffer to decode from
     * @returns FlashReadRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decodeDelimited(reader: $protobuf.Reader | Uint8Array): blerpc.FlashReadRequest;

    /**
     * Verifies a FlashReadRequest message.
     * @param message Plain object to verify
     * @returns `null` if valid, otherwise the reason why it is not
     */
    public static verify(message: { [k: string]: any }): string | null;

    /**
     * Creates a FlashReadRequest message from a plain object. Also converts values to their respective internal types.
     * @param object Plain object
     * @returns FlashReadRequest
     */
    public static fromObject(object: { [k: string]: any }): blerpc.FlashReadRequest;

    /**
     * Creates a plain object from a FlashReadRequest message. Also converts values to other types if specified.
     * @param message FlashReadRequest
     * @param [options] Conversion options
     * @returns Plain object
     */
    public static toObject(
      message: blerpc.FlashReadRequest,
      options?: $protobuf.IConversionOptions,
    ): { [k: string]: any };

    /**
     * Converts this FlashReadRequest to JSON.
     * @returns JSON object
     */
    public toJSON(): { [k: string]: any };

    /**
     * Gets the default type url for FlashReadRequest
     * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
     * @returns The default type url
     */
    public static getTypeUrl(typeUrlPrefix?: string): string;
  }

  /** Properties of a FlashReadResponse. */
  interface IFlashReadResponse {
    /** FlashReadResponse address */
    address?: number | null;

    /** FlashReadResponse data */
    data?: Uint8Array | null;
  }

  /** Represents a FlashReadResponse. */
  class FlashReadResponse implements IFlashReadResponse {
    /**
     * Constructs a new FlashReadResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: blerpc.IFlashReadResponse);

    /** FlashReadResponse address. */
    public address: number;

    /** FlashReadResponse data. */
    public data: Uint8Array;

    /**
     * Creates a new FlashReadResponse instance using the specified properties.
     * @param [properties] Properties to set
     * @returns FlashReadResponse instance
     */
    public static create(properties?: blerpc.IFlashReadResponse): blerpc.FlashReadResponse;

    /**
     * Encodes the specified FlashReadResponse message. Does not implicitly {@link blerpc.FlashReadResponse.verify|verify} messages.
     * @param message FlashReadResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(
      message: blerpc.IFlashReadResponse,
      writer?: $protobuf.Writer,
    ): $protobuf.Writer;

    /**
     * Encodes the specified FlashReadResponse message, length delimited. Does not implicitly {@link blerpc.FlashReadResponse.verify|verify} messages.
     * @param message FlashReadResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encodeDelimited(
      message: blerpc.IFlashReadResponse,
      writer?: $protobuf.Writer,
    ): $protobuf.Writer;

    /**
     * Decodes a FlashReadResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns FlashReadResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(
      reader: $protobuf.Reader | Uint8Array,
      length?: number,
    ): blerpc.FlashReadResponse;

    /**
     * Decodes a FlashReadResponse message from the specified reader or buffer, length delimited.
     * @param reader Reader or buffer to decode from
     * @returns FlashReadResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decodeDelimited(reader: $protobuf.Reader | Uint8Array): blerpc.FlashReadResponse;

    /**
     * Verifies a FlashReadResponse message.
     * @param message Plain object to verify
     * @returns `null` if valid, otherwise the reason why it is not
     */
    public static verify(message: { [k: string]: any }): string | null;

    /**
     * Creates a FlashReadResponse message from a plain object. Also converts values to their respective internal types.
     * @param object Plain object
     * @returns FlashReadResponse
     */
    public static fromObject(object: { [k: string]: any }): blerpc.FlashReadResponse;

    /**
     * Creates a plain object from a FlashReadResponse message. Also converts values to other types if specified.
     * @param message FlashReadResponse
     * @param [options] Conversion options
     * @returns Plain object
     */
    public static toObject(
      message: blerpc.FlashReadResponse,
      options?: $protobuf.IConversionOptions,
    ): { [k: string]: any };

    /**
     * Converts this FlashReadResponse to JSON.
     * @returns JSON object
     */
    public toJSON(): { [k: string]: any };

    /**
     * Gets the default type url for FlashReadResponse
     * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
     * @returns The default type url
     */
    public static getTypeUrl(typeUrlPrefix?: string): string;
  }

  /** Properties of a DataWriteRequest. */
  interface IDataWriteRequest {
    /** DataWriteRequest data */
    data?: Uint8Array | null;
  }

  /** Represents a DataWriteRequest. */
  class DataWriteRequest implements IDataWriteRequest {
    /**
     * Constructs a new DataWriteRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: blerpc.IDataWriteRequest);

    /** DataWriteRequest data. */
    public data: Uint8Array;

    /**
     * Creates a new DataWriteRequest instance using the specified properties.
     * @param [properties] Properties to set
     * @returns DataWriteRequest instance
     */
    public static create(properties?: blerpc.IDataWriteRequest): blerpc.DataWriteRequest;

    /**
     * Encodes the specified DataWriteRequest message. Does not implicitly {@link blerpc.DataWriteRequest.verify|verify} messages.
     * @param message DataWriteRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(
      message: blerpc.IDataWriteRequest,
      writer?: $protobuf.Writer,
    ): $protobuf.Writer;

    /**
     * Encodes the specified DataWriteRequest message, length delimited. Does not implicitly {@link blerpc.DataWriteRequest.verify|verify} messages.
     * @param message DataWriteRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encodeDelimited(
      message: blerpc.IDataWriteRequest,
      writer?: $protobuf.Writer,
    ): $protobuf.Writer;

    /**
     * Decodes a DataWriteRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns DataWriteRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(
      reader: $protobuf.Reader | Uint8Array,
      length?: number,
    ): blerpc.DataWriteRequest;

    /**
     * Decodes a DataWriteRequest message from the specified reader or buffer, length delimited.
     * @param reader Reader or buffer to decode from
     * @returns DataWriteRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decodeDelimited(reader: $protobuf.Reader | Uint8Array): blerpc.DataWriteRequest;

    /**
     * Verifies a DataWriteRequest message.
     * @param message Plain object to verify
     * @returns `null` if valid, otherwise the reason why it is not
     */
    public static verify(message: { [k: string]: any }): string | null;

    /**
     * Creates a DataWriteRequest message from a plain object. Also converts values to their respective internal types.
     * @param object Plain object
     * @returns DataWriteRequest
     */
    public static fromObject(object: { [k: string]: any }): blerpc.DataWriteRequest;

    /**
     * Creates a plain object from a DataWriteRequest message. Also converts values to other types if specified.
     * @param message DataWriteRequest
     * @param [options] Conversion options
     * @returns Plain object
     */
    public static toObject(
      message: blerpc.DataWriteRequest,
      options?: $protobuf.IConversionOptions,
    ): { [k: string]: any };

    /**
     * Converts this DataWriteRequest to JSON.
     * @returns JSON object
     */
    public toJSON(): { [k: string]: any };

    /**
     * Gets the default type url for DataWriteRequest
     * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
     * @returns The default type url
     */
    public static getTypeUrl(typeUrlPrefix?: string): string;
  }

  /** Properties of a DataWriteResponse. */
  interface IDataWriteResponse {
    /** DataWriteResponse length */
    length?: number | null;
  }

  /** Represents a DataWriteResponse. */
  class DataWriteResponse implements IDataWriteResponse {
    /**
     * Constructs a new DataWriteResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: blerpc.IDataWriteResponse);

    /** DataWriteResponse length. */
    public length: number;

    /**
     * Creates a new DataWriteResponse instance using the specified properties.
     * @param [properties] Properties to set
     * @returns DataWriteResponse instance
     */
    public static create(properties?: blerpc.IDataWriteResponse): blerpc.DataWriteResponse;

    /**
     * Encodes the specified DataWriteResponse message. Does not implicitly {@link blerpc.DataWriteResponse.verify|verify} messages.
     * @param message DataWriteResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(
      message: blerpc.IDataWriteResponse,
      writer?: $protobuf.Writer,
    ): $protobuf.Writer;

    /**
     * Encodes the specified DataWriteResponse message, length delimited. Does not implicitly {@link blerpc.DataWriteResponse.verify|verify} messages.
     * @param message DataWriteResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encodeDelimited(
      message: blerpc.IDataWriteResponse,
      writer?: $protobuf.Writer,
    ): $protobuf.Writer;

    /**
     * Decodes a DataWriteResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns DataWriteResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(
      reader: $protobuf.Reader | Uint8Array,
      length?: number,
    ): blerpc.DataWriteResponse;

    /**
     * Decodes a DataWriteResponse message from the specified reader or buffer, length delimited.
     * @param reader Reader or buffer to decode from
     * @returns DataWriteResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decodeDelimited(reader: $protobuf.Reader | Uint8Array): blerpc.DataWriteResponse;

    /**
     * Verifies a DataWriteResponse message.
     * @param message Plain object to verify
     * @returns `null` if valid, otherwise the reason why it is not
     */
    public static verify(message: { [k: string]: any }): string | null;

    /**
     * Creates a DataWriteResponse message from a plain object. Also converts values to their respective internal types.
     * @param object Plain object
     * @returns DataWriteResponse
     */
    public static fromObject(object: { [k: string]: any }): blerpc.DataWriteResponse;

    /**
     * Creates a plain object from a DataWriteResponse message. Also converts values to other types if specified.
     * @param message DataWriteResponse
     * @param [options] Conversion options
     * @returns Plain object
     */
    public static toObject(
      message: blerpc.DataWriteResponse,
      options?: $protobuf.IConversionOptions,
    ): { [k: string]: any };

    /**
     * Converts this DataWriteResponse to JSON.
     * @returns JSON object
     */
    public toJSON(): { [k: string]: any };

    /**
     * Gets the default type url for DataWriteResponse
     * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
     * @returns The default type url
     */
    public static getTypeUrl(typeUrlPrefix?: string): string;
  }

  /** Properties of a CounterStreamRequest. */
  interface ICounterStreamRequest {
    /** CounterStreamRequest count */
    count?: number | null;
  }

  /** Represents a CounterStreamRequest. */
  class CounterStreamRequest implements ICounterStreamRequest {
    /**
     * Constructs a new CounterStreamRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: blerpc.ICounterStreamRequest);

    /** CounterStreamRequest count. */
    public count: number;

    /**
     * Creates a new CounterStreamRequest instance using the specified properties.
     * @param [properties] Properties to set
     * @returns CounterStreamRequest instance
     */
    public static create(properties?: blerpc.ICounterStreamRequest): blerpc.CounterStreamRequest;

    /**
     * Encodes the specified CounterStreamRequest message. Does not implicitly {@link blerpc.CounterStreamRequest.verify|verify} messages.
     * @param message CounterStreamRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(
      message: blerpc.ICounterStreamRequest,
      writer?: $protobuf.Writer,
    ): $protobuf.Writer;

    /**
     * Encodes the specified CounterStreamRequest message, length delimited. Does not implicitly {@link blerpc.CounterStreamRequest.verify|verify} messages.
     * @param message CounterStreamRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encodeDelimited(
      message: blerpc.ICounterStreamRequest,
      writer?: $protobuf.Writer,
    ): $protobuf.Writer;

    /**
     * Decodes a CounterStreamRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns CounterStreamRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(
      reader: $protobuf.Reader | Uint8Array,
      length?: number,
    ): blerpc.CounterStreamRequest;

    /**
     * Decodes a CounterStreamRequest message from the specified reader or buffer, length delimited.
     * @param reader Reader or buffer to decode from
     * @returns CounterStreamRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decodeDelimited(
      reader: $protobuf.Reader | Uint8Array,
    ): blerpc.CounterStreamRequest;

    /**
     * Verifies a CounterStreamRequest message.
     * @param message Plain object to verify
     * @returns `null` if valid, otherwise the reason why it is not
     */
    public static verify(message: { [k: string]: any }): string | null;

    /**
     * Creates a CounterStreamRequest message from a plain object. Also converts values to their respective internal types.
     * @param object Plain object
     * @returns CounterStreamRequest
     */
    public static fromObject(object: { [k: string]: any }): blerpc.CounterStreamRequest;

    /**
     * Creates a plain object from a CounterStreamRequest message. Also converts values to other types if specified.
     * @param message CounterStreamRequest
     * @param [options] Conversion options
     * @returns Plain object
     */
    public static toObject(
      message: blerpc.CounterStreamRequest,
      options?: $protobuf.IConversionOptions,
    ): { [k: string]: any };

    /**
     * Converts this CounterStreamRequest to JSON.
     * @returns JSON object
     */
    public toJSON(): { [k: string]: any };

    /**
     * Gets the default type url for CounterStreamRequest
     * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
     * @returns The default type url
     */
    public static getTypeUrl(typeUrlPrefix?: string): string;
  }

  /** Properties of a CounterStreamResponse. */
  interface ICounterStreamResponse {
    /** CounterStreamResponse seq */
    seq?: number | null;

    /** CounterStreamResponse value */
    value?: number | null;
  }

  /** Represents a CounterStreamResponse. */
  class CounterStreamResponse implements ICounterStreamResponse {
    /**
     * Constructs a new CounterStreamResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: blerpc.ICounterStreamResponse);

    /** CounterStreamResponse seq. */
    public seq: number;

    /** CounterStreamResponse value. */
    public value: number;

    /**
     * Creates a new CounterStreamResponse instance using the specified properties.
     * @param [properties] Properties to set
     * @returns CounterStreamResponse instance
     */
    public static create(properties?: blerpc.ICounterStreamResponse): blerpc.CounterStreamResponse;

    /**
     * Encodes the specified CounterStreamResponse message. Does not implicitly {@link blerpc.CounterStreamResponse.verify|verify} messages.
     * @param message CounterStreamResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(
      message: blerpc.ICounterStreamResponse,
      writer?: $protobuf.Writer,
    ): $protobuf.Writer;

    /**
     * Encodes the specified CounterStreamResponse message, length delimited. Does not implicitly {@link blerpc.CounterStreamResponse.verify|verify} messages.
     * @param message CounterStreamResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encodeDelimited(
      message: blerpc.ICounterStreamResponse,
      writer?: $protobuf.Writer,
    ): $protobuf.Writer;

    /**
     * Decodes a CounterStreamResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns CounterStreamResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(
      reader: $protobuf.Reader | Uint8Array,
      length?: number,
    ): blerpc.CounterStreamResponse;

    /**
     * Decodes a CounterStreamResponse message from the specified reader or buffer, length delimited.
     * @param reader Reader or buffer to decode from
     * @returns CounterStreamResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decodeDelimited(
      reader: $protobuf.Reader | Uint8Array,
    ): blerpc.CounterStreamResponse;

    /**
     * Verifies a CounterStreamResponse message.
     * @param message Plain object to verify
     * @returns `null` if valid, otherwise the reason why it is not
     */
    public static verify(message: { [k: string]: any }): string | null;

    /**
     * Creates a CounterStreamResponse message from a plain object. Also converts values to their respective internal types.
     * @param object Plain object
     * @returns CounterStreamResponse
     */
    public static fromObject(object: { [k: string]: any }): blerpc.CounterStreamResponse;

    /**
     * Creates a plain object from a CounterStreamResponse message. Also converts values to other types if specified.
     * @param message CounterStreamResponse
     * @param [options] Conversion options
     * @returns Plain object
     */
    public static toObject(
      message: blerpc.CounterStreamResponse,
      options?: $protobuf.IConversionOptions,
    ): { [k: string]: any };

    /**
     * Converts this CounterStreamResponse to JSON.
     * @returns JSON object
     */
    public toJSON(): { [k: string]: any };

    /**
     * Gets the default type url for CounterStreamResponse
     * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
     * @returns The default type url
     */
    public static getTypeUrl(typeUrlPrefix?: string): string;
  }

  /** Properties of a CounterUploadRequest. */
  interface ICounterUploadRequest {
    /** CounterUploadRequest seq */
    seq?: number | null;

    /** CounterUploadRequest value */
    value?: number | null;
  }

  /** Represents a CounterUploadRequest. */
  class CounterUploadRequest implements ICounterUploadRequest {
    /**
     * Constructs a new CounterUploadRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: blerpc.ICounterUploadRequest);

    /** CounterUploadRequest seq. */
    public seq: number;

    /** CounterUploadRequest value. */
    public value: number;

    /**
     * Creates a new CounterUploadRequest instance using the specified properties.
     * @param [properties] Properties to set
     * @returns CounterUploadRequest instance
     */
    public static create(properties?: blerpc.ICounterUploadRequest): blerpc.CounterUploadRequest;

    /**
     * Encodes the specified CounterUploadRequest message. Does not implicitly {@link blerpc.CounterUploadRequest.verify|verify} messages.
     * @param message CounterUploadRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(
      message: blerpc.ICounterUploadRequest,
      writer?: $protobuf.Writer,
    ): $protobuf.Writer;

    /**
     * Encodes the specified CounterUploadRequest message, length delimited. Does not implicitly {@link blerpc.CounterUploadRequest.verify|verify} messages.
     * @param message CounterUploadRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encodeDelimited(
      message: blerpc.ICounterUploadRequest,
      writer?: $protobuf.Writer,
    ): $protobuf.Writer;

    /**
     * Decodes a CounterUploadRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns CounterUploadRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(
      reader: $protobuf.Reader | Uint8Array,
      length?: number,
    ): blerpc.CounterUploadRequest;

    /**
     * Decodes a CounterUploadRequest message from the specified reader or buffer, length delimited.
     * @param reader Reader or buffer to decode from
     * @returns CounterUploadRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decodeDelimited(
      reader: $protobuf.Reader | Uint8Array,
    ): blerpc.CounterUploadRequest;

    /**
     * Verifies a CounterUploadRequest message.
     * @param message Plain object to verify
     * @returns `null` if valid, otherwise the reason why it is not
     */
    public static verify(message: { [k: string]: any }): string | null;

    /**
     * Creates a CounterUploadRequest message from a plain object. Also converts values to their respective internal types.
     * @param object Plain object
     * @returns CounterUploadRequest
     */
    public static fromObject(object: { [k: string]: any }): blerpc.CounterUploadRequest;

    /**
     * Creates a plain object from a CounterUploadRequest message. Also converts values to other types if specified.
     * @param message CounterUploadRequest
     * @param [options] Conversion options
     * @returns Plain object
     */
    public static toObject(
      message: blerpc.CounterUploadRequest,
      options?: $protobuf.IConversionOptions,
    ): { [k: string]: any };

    /**
     * Converts this CounterUploadRequest to JSON.
     * @returns JSON object
     */
    public toJSON(): { [k: string]: any };

    /**
     * Gets the default type url for CounterUploadRequest
     * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
     * @returns The default type url
     */
    public static getTypeUrl(typeUrlPrefix?: string): string;
  }

  /** Properties of a CounterUploadResponse. */
  interface ICounterUploadResponse {
    /** CounterUploadResponse receivedCount */
    receivedCount?: number | null;
  }

  /** Represents a CounterUploadResponse. */
  class CounterUploadResponse implements ICounterUploadResponse {
    /**
     * Constructs a new CounterUploadResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: blerpc.ICounterUploadResponse);

    /** CounterUploadResponse receivedCount. */
    public receivedCount: number;

    /**
     * Creates a new CounterUploadResponse instance using the specified properties.
     * @param [properties] Properties to set
     * @returns CounterUploadResponse instance
     */
    public static create(properties?: blerpc.ICounterUploadResponse): blerpc.CounterUploadResponse;

    /**
     * Encodes the specified CounterUploadResponse message. Does not implicitly {@link blerpc.CounterUploadResponse.verify|verify} messages.
     * @param message CounterUploadResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(
      message: blerpc.ICounterUploadResponse,
      writer?: $protobuf.Writer,
    ): $protobuf.Writer;

    /**
     * Encodes the specified CounterUploadResponse message, length delimited. Does not implicitly {@link blerpc.CounterUploadResponse.verify|verify} messages.
     * @param message CounterUploadResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encodeDelimited(
      message: blerpc.ICounterUploadResponse,
      writer?: $protobuf.Writer,
    ): $protobuf.Writer;

    /**
     * Decodes a CounterUploadResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns CounterUploadResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(
      reader: $protobuf.Reader | Uint8Array,
      length?: number,
    ): blerpc.CounterUploadResponse;

    /**
     * Decodes a CounterUploadResponse message from the specified reader or buffer, length delimited.
     * @param reader Reader or buffer to decode from
     * @returns CounterUploadResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decodeDelimited(
      reader: $protobuf.Reader | Uint8Array,
    ): blerpc.CounterUploadResponse;

    /**
     * Verifies a CounterUploadResponse message.
     * @param message Plain object to verify
     * @returns `null` if valid, otherwise the reason why it is not
     */
    public static verify(message: { [k: string]: any }): string | null;

    /**
     * Creates a CounterUploadResponse message from a plain object. Also converts values to their respective internal types.
     * @param object Plain object
     * @returns CounterUploadResponse
     */
    public static fromObject(object: { [k: string]: any }): blerpc.CounterUploadResponse;

    /**
     * Creates a plain object from a CounterUploadResponse message. Also converts values to other types if specified.
     * @param message CounterUploadResponse
     * @param [options] Conversion options
     * @returns Plain object
     */
    public static toObject(
      message: blerpc.CounterUploadResponse,
      options?: $protobuf.IConversionOptions,
    ): { [k: string]: any };

    /**
     * Converts this CounterUploadResponse to JSON.
     * @returns JSON object
     */
    public toJSON(): { [k: string]: any };

    /**
     * Gets the default type url for CounterUploadResponse
     * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
     * @returns The default type url
     */
    public static getTypeUrl(typeUrlPrefix?: string): string;
  }
}
