/*eslint-disable block-scoped-var, id-length, no-control-regex, no-magic-numbers, no-prototype-builtins, no-redeclare, no-shadow, no-var, sort-vars*/
"use strict";

var $protobuf = require("protobufjs/minimal");

// Common aliases
var $Reader = $protobuf.Reader, $Writer = $protobuf.Writer, $util = $protobuf.util;

// Exported root namespace
var $root = $protobuf.roots["default"] || ($protobuf.roots["default"] = {});

$root.blerpc = (function() {

    /**
     * Namespace blerpc.
     * @exports blerpc
     * @namespace
     */
    var blerpc = {};

    blerpc.EchoRequest = (function() {

        /**
         * Properties of an EchoRequest.
         * @memberof blerpc
         * @interface IEchoRequest
         * @property {string|null} [message] EchoRequest message
         */

        /**
         * Constructs a new EchoRequest.
         * @memberof blerpc
         * @classdesc Represents an EchoRequest.
         * @implements IEchoRequest
         * @constructor
         * @param {blerpc.IEchoRequest=} [properties] Properties to set
         */
        function EchoRequest(properties) {
            if (properties)
                for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * EchoRequest message.
         * @member {string} message
         * @memberof blerpc.EchoRequest
         * @instance
         */
        EchoRequest.prototype.message = "";

        /**
         * Creates a new EchoRequest instance using the specified properties.
         * @function create
         * @memberof blerpc.EchoRequest
         * @static
         * @param {blerpc.IEchoRequest=} [properties] Properties to set
         * @returns {blerpc.EchoRequest} EchoRequest instance
         */
        EchoRequest.create = function create(properties) {
            return new EchoRequest(properties);
        };

        /**
         * Encodes the specified EchoRequest message. Does not implicitly {@link blerpc.EchoRequest.verify|verify} messages.
         * @function encode
         * @memberof blerpc.EchoRequest
         * @static
         * @param {blerpc.IEchoRequest} message EchoRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        EchoRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.message != null && Object.hasOwnProperty.call(message, "message"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.message);
            return writer;
        };

        /**
         * Encodes the specified EchoRequest message, length delimited. Does not implicitly {@link blerpc.EchoRequest.verify|verify} messages.
         * @function encodeDelimited
         * @memberof blerpc.EchoRequest
         * @static
         * @param {blerpc.IEchoRequest} message EchoRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        EchoRequest.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes an EchoRequest message from the specified reader or buffer.
         * @function decode
         * @memberof blerpc.EchoRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {blerpc.EchoRequest} EchoRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        EchoRequest.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            var end = length === undefined ? reader.len : reader.pos + length, message = new $root.blerpc.EchoRequest();
            while (reader.pos < end) {
                var tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.message = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes an EchoRequest message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof blerpc.EchoRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {blerpc.EchoRequest} EchoRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        EchoRequest.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies an EchoRequest message.
         * @function verify
         * @memberof blerpc.EchoRequest
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        EchoRequest.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.message != null && message.hasOwnProperty("message"))
                if (!$util.isString(message.message))
                    return "message: string expected";
            return null;
        };

        /**
         * Creates an EchoRequest message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof blerpc.EchoRequest
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {blerpc.EchoRequest} EchoRequest
         */
        EchoRequest.fromObject = function fromObject(object) {
            if (object instanceof $root.blerpc.EchoRequest)
                return object;
            var message = new $root.blerpc.EchoRequest();
            if (object.message != null)
                message.message = String(object.message);
            return message;
        };

        /**
         * Creates a plain object from an EchoRequest message. Also converts values to other types if specified.
         * @function toObject
         * @memberof blerpc.EchoRequest
         * @static
         * @param {blerpc.EchoRequest} message EchoRequest
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        EchoRequest.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            var object = {};
            if (options.defaults)
                object.message = "";
            if (message.message != null && message.hasOwnProperty("message"))
                object.message = message.message;
            return object;
        };

        /**
         * Converts this EchoRequest to JSON.
         * @function toJSON
         * @memberof blerpc.EchoRequest
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        EchoRequest.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for EchoRequest
         * @function getTypeUrl
         * @memberof blerpc.EchoRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        EchoRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/blerpc.EchoRequest";
        };

        return EchoRequest;
    })();

    blerpc.EchoResponse = (function() {

        /**
         * Properties of an EchoResponse.
         * @memberof blerpc
         * @interface IEchoResponse
         * @property {string|null} [message] EchoResponse message
         */

        /**
         * Constructs a new EchoResponse.
         * @memberof blerpc
         * @classdesc Represents an EchoResponse.
         * @implements IEchoResponse
         * @constructor
         * @param {blerpc.IEchoResponse=} [properties] Properties to set
         */
        function EchoResponse(properties) {
            if (properties)
                for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * EchoResponse message.
         * @member {string} message
         * @memberof blerpc.EchoResponse
         * @instance
         */
        EchoResponse.prototype.message = "";

        /**
         * Creates a new EchoResponse instance using the specified properties.
         * @function create
         * @memberof blerpc.EchoResponse
         * @static
         * @param {blerpc.IEchoResponse=} [properties] Properties to set
         * @returns {blerpc.EchoResponse} EchoResponse instance
         */
        EchoResponse.create = function create(properties) {
            return new EchoResponse(properties);
        };

        /**
         * Encodes the specified EchoResponse message. Does not implicitly {@link blerpc.EchoResponse.verify|verify} messages.
         * @function encode
         * @memberof blerpc.EchoResponse
         * @static
         * @param {blerpc.IEchoResponse} message EchoResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        EchoResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.message != null && Object.hasOwnProperty.call(message, "message"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.message);
            return writer;
        };

        /**
         * Encodes the specified EchoResponse message, length delimited. Does not implicitly {@link blerpc.EchoResponse.verify|verify} messages.
         * @function encodeDelimited
         * @memberof blerpc.EchoResponse
         * @static
         * @param {blerpc.IEchoResponse} message EchoResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        EchoResponse.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes an EchoResponse message from the specified reader or buffer.
         * @function decode
         * @memberof blerpc.EchoResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {blerpc.EchoResponse} EchoResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        EchoResponse.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            var end = length === undefined ? reader.len : reader.pos + length, message = new $root.blerpc.EchoResponse();
            while (reader.pos < end) {
                var tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.message = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes an EchoResponse message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof blerpc.EchoResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {blerpc.EchoResponse} EchoResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        EchoResponse.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies an EchoResponse message.
         * @function verify
         * @memberof blerpc.EchoResponse
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        EchoResponse.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.message != null && message.hasOwnProperty("message"))
                if (!$util.isString(message.message))
                    return "message: string expected";
            return null;
        };

        /**
         * Creates an EchoResponse message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof blerpc.EchoResponse
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {blerpc.EchoResponse} EchoResponse
         */
        EchoResponse.fromObject = function fromObject(object) {
            if (object instanceof $root.blerpc.EchoResponse)
                return object;
            var message = new $root.blerpc.EchoResponse();
            if (object.message != null)
                message.message = String(object.message);
            return message;
        };

        /**
         * Creates a plain object from an EchoResponse message. Also converts values to other types if specified.
         * @function toObject
         * @memberof blerpc.EchoResponse
         * @static
         * @param {blerpc.EchoResponse} message EchoResponse
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        EchoResponse.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            var object = {};
            if (options.defaults)
                object.message = "";
            if (message.message != null && message.hasOwnProperty("message"))
                object.message = message.message;
            return object;
        };

        /**
         * Converts this EchoResponse to JSON.
         * @function toJSON
         * @memberof blerpc.EchoResponse
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        EchoResponse.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for EchoResponse
         * @function getTypeUrl
         * @memberof blerpc.EchoResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        EchoResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/blerpc.EchoResponse";
        };

        return EchoResponse;
    })();

    blerpc.FlashReadRequest = (function() {

        /**
         * Properties of a FlashReadRequest.
         * @memberof blerpc
         * @interface IFlashReadRequest
         * @property {number|null} [address] FlashReadRequest address
         * @property {number|null} [length] FlashReadRequest length
         */

        /**
         * Constructs a new FlashReadRequest.
         * @memberof blerpc
         * @classdesc Represents a FlashReadRequest.
         * @implements IFlashReadRequest
         * @constructor
         * @param {blerpc.IFlashReadRequest=} [properties] Properties to set
         */
        function FlashReadRequest(properties) {
            if (properties)
                for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * FlashReadRequest address.
         * @member {number} address
         * @memberof blerpc.FlashReadRequest
         * @instance
         */
        FlashReadRequest.prototype.address = 0;

        /**
         * FlashReadRequest length.
         * @member {number} length
         * @memberof blerpc.FlashReadRequest
         * @instance
         */
        FlashReadRequest.prototype.length = 0;

        /**
         * Creates a new FlashReadRequest instance using the specified properties.
         * @function create
         * @memberof blerpc.FlashReadRequest
         * @static
         * @param {blerpc.IFlashReadRequest=} [properties] Properties to set
         * @returns {blerpc.FlashReadRequest} FlashReadRequest instance
         */
        FlashReadRequest.create = function create(properties) {
            return new FlashReadRequest(properties);
        };

        /**
         * Encodes the specified FlashReadRequest message. Does not implicitly {@link blerpc.FlashReadRequest.verify|verify} messages.
         * @function encode
         * @memberof blerpc.FlashReadRequest
         * @static
         * @param {blerpc.IFlashReadRequest} message FlashReadRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        FlashReadRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.address != null && Object.hasOwnProperty.call(message, "address"))
                writer.uint32(/* id 1, wireType 0 =*/8).uint32(message.address);
            if (message.length != null && Object.hasOwnProperty.call(message, "length"))
                writer.uint32(/* id 2, wireType 0 =*/16).uint32(message.length);
            return writer;
        };

        /**
         * Encodes the specified FlashReadRequest message, length delimited. Does not implicitly {@link blerpc.FlashReadRequest.verify|verify} messages.
         * @function encodeDelimited
         * @memberof blerpc.FlashReadRequest
         * @static
         * @param {blerpc.IFlashReadRequest} message FlashReadRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        FlashReadRequest.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a FlashReadRequest message from the specified reader or buffer.
         * @function decode
         * @memberof blerpc.FlashReadRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {blerpc.FlashReadRequest} FlashReadRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        FlashReadRequest.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            var end = length === undefined ? reader.len : reader.pos + length, message = new $root.blerpc.FlashReadRequest();
            while (reader.pos < end) {
                var tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.address = reader.uint32();
                        break;
                    }
                case 2: {
                        message.length = reader.uint32();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a FlashReadRequest message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof blerpc.FlashReadRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {blerpc.FlashReadRequest} FlashReadRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        FlashReadRequest.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a FlashReadRequest message.
         * @function verify
         * @memberof blerpc.FlashReadRequest
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        FlashReadRequest.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.address != null && message.hasOwnProperty("address"))
                if (!$util.isInteger(message.address))
                    return "address: integer expected";
            if (message.length != null && message.hasOwnProperty("length"))
                if (!$util.isInteger(message.length))
                    return "length: integer expected";
            return null;
        };

        /**
         * Creates a FlashReadRequest message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof blerpc.FlashReadRequest
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {blerpc.FlashReadRequest} FlashReadRequest
         */
        FlashReadRequest.fromObject = function fromObject(object) {
            if (object instanceof $root.blerpc.FlashReadRequest)
                return object;
            var message = new $root.blerpc.FlashReadRequest();
            if (object.address != null)
                message.address = object.address >>> 0;
            if (object.length != null)
                message.length = object.length >>> 0;
            return message;
        };

        /**
         * Creates a plain object from a FlashReadRequest message. Also converts values to other types if specified.
         * @function toObject
         * @memberof blerpc.FlashReadRequest
         * @static
         * @param {blerpc.FlashReadRequest} message FlashReadRequest
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        FlashReadRequest.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            var object = {};
            if (options.defaults) {
                object.address = 0;
                object.length = 0;
            }
            if (message.address != null && message.hasOwnProperty("address"))
                object.address = message.address;
            if (message.length != null && message.hasOwnProperty("length"))
                object.length = message.length;
            return object;
        };

        /**
         * Converts this FlashReadRequest to JSON.
         * @function toJSON
         * @memberof blerpc.FlashReadRequest
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        FlashReadRequest.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for FlashReadRequest
         * @function getTypeUrl
         * @memberof blerpc.FlashReadRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        FlashReadRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/blerpc.FlashReadRequest";
        };

        return FlashReadRequest;
    })();

    blerpc.FlashReadResponse = (function() {

        /**
         * Properties of a FlashReadResponse.
         * @memberof blerpc
         * @interface IFlashReadResponse
         * @property {number|null} [address] FlashReadResponse address
         * @property {Uint8Array|null} [data] FlashReadResponse data
         */

        /**
         * Constructs a new FlashReadResponse.
         * @memberof blerpc
         * @classdesc Represents a FlashReadResponse.
         * @implements IFlashReadResponse
         * @constructor
         * @param {blerpc.IFlashReadResponse=} [properties] Properties to set
         */
        function FlashReadResponse(properties) {
            if (properties)
                for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * FlashReadResponse address.
         * @member {number} address
         * @memberof blerpc.FlashReadResponse
         * @instance
         */
        FlashReadResponse.prototype.address = 0;

        /**
         * FlashReadResponse data.
         * @member {Uint8Array} data
         * @memberof blerpc.FlashReadResponse
         * @instance
         */
        FlashReadResponse.prototype.data = $util.newBuffer([]);

        /**
         * Creates a new FlashReadResponse instance using the specified properties.
         * @function create
         * @memberof blerpc.FlashReadResponse
         * @static
         * @param {blerpc.IFlashReadResponse=} [properties] Properties to set
         * @returns {blerpc.FlashReadResponse} FlashReadResponse instance
         */
        FlashReadResponse.create = function create(properties) {
            return new FlashReadResponse(properties);
        };

        /**
         * Encodes the specified FlashReadResponse message. Does not implicitly {@link blerpc.FlashReadResponse.verify|verify} messages.
         * @function encode
         * @memberof blerpc.FlashReadResponse
         * @static
         * @param {blerpc.IFlashReadResponse} message FlashReadResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        FlashReadResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.address != null && Object.hasOwnProperty.call(message, "address"))
                writer.uint32(/* id 1, wireType 0 =*/8).uint32(message.address);
            if (message.data != null && Object.hasOwnProperty.call(message, "data"))
                writer.uint32(/* id 2, wireType 2 =*/18).bytes(message.data);
            return writer;
        };

        /**
         * Encodes the specified FlashReadResponse message, length delimited. Does not implicitly {@link blerpc.FlashReadResponse.verify|verify} messages.
         * @function encodeDelimited
         * @memberof blerpc.FlashReadResponse
         * @static
         * @param {blerpc.IFlashReadResponse} message FlashReadResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        FlashReadResponse.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a FlashReadResponse message from the specified reader or buffer.
         * @function decode
         * @memberof blerpc.FlashReadResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {blerpc.FlashReadResponse} FlashReadResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        FlashReadResponse.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            var end = length === undefined ? reader.len : reader.pos + length, message = new $root.blerpc.FlashReadResponse();
            while (reader.pos < end) {
                var tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.address = reader.uint32();
                        break;
                    }
                case 2: {
                        message.data = reader.bytes();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a FlashReadResponse message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof blerpc.FlashReadResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {blerpc.FlashReadResponse} FlashReadResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        FlashReadResponse.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a FlashReadResponse message.
         * @function verify
         * @memberof blerpc.FlashReadResponse
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        FlashReadResponse.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.address != null && message.hasOwnProperty("address"))
                if (!$util.isInteger(message.address))
                    return "address: integer expected";
            if (message.data != null && message.hasOwnProperty("data"))
                if (!(message.data && typeof message.data.length === "number" || $util.isString(message.data)))
                    return "data: buffer expected";
            return null;
        };

        /**
         * Creates a FlashReadResponse message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof blerpc.FlashReadResponse
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {blerpc.FlashReadResponse} FlashReadResponse
         */
        FlashReadResponse.fromObject = function fromObject(object) {
            if (object instanceof $root.blerpc.FlashReadResponse)
                return object;
            var message = new $root.blerpc.FlashReadResponse();
            if (object.address != null)
                message.address = object.address >>> 0;
            if (object.data != null)
                if (typeof object.data === "string")
                    $util.base64.decode(object.data, message.data = $util.newBuffer($util.base64.length(object.data)), 0);
                else if (object.data.length >= 0)
                    message.data = object.data;
            return message;
        };

        /**
         * Creates a plain object from a FlashReadResponse message. Also converts values to other types if specified.
         * @function toObject
         * @memberof blerpc.FlashReadResponse
         * @static
         * @param {blerpc.FlashReadResponse} message FlashReadResponse
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        FlashReadResponse.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            var object = {};
            if (options.defaults) {
                object.address = 0;
                if (options.bytes === String)
                    object.data = "";
                else {
                    object.data = [];
                    if (options.bytes !== Array)
                        object.data = $util.newBuffer(object.data);
                }
            }
            if (message.address != null && message.hasOwnProperty("address"))
                object.address = message.address;
            if (message.data != null && message.hasOwnProperty("data"))
                object.data = options.bytes === String ? $util.base64.encode(message.data, 0, message.data.length) : options.bytes === Array ? Array.prototype.slice.call(message.data) : message.data;
            return object;
        };

        /**
         * Converts this FlashReadResponse to JSON.
         * @function toJSON
         * @memberof blerpc.FlashReadResponse
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        FlashReadResponse.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for FlashReadResponse
         * @function getTypeUrl
         * @memberof blerpc.FlashReadResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        FlashReadResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/blerpc.FlashReadResponse";
        };

        return FlashReadResponse;
    })();

    blerpc.DataWriteRequest = (function() {

        /**
         * Properties of a DataWriteRequest.
         * @memberof blerpc
         * @interface IDataWriteRequest
         * @property {Uint8Array|null} [data] DataWriteRequest data
         */

        /**
         * Constructs a new DataWriteRequest.
         * @memberof blerpc
         * @classdesc Represents a DataWriteRequest.
         * @implements IDataWriteRequest
         * @constructor
         * @param {blerpc.IDataWriteRequest=} [properties] Properties to set
         */
        function DataWriteRequest(properties) {
            if (properties)
                for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * DataWriteRequest data.
         * @member {Uint8Array} data
         * @memberof blerpc.DataWriteRequest
         * @instance
         */
        DataWriteRequest.prototype.data = $util.newBuffer([]);

        /**
         * Creates a new DataWriteRequest instance using the specified properties.
         * @function create
         * @memberof blerpc.DataWriteRequest
         * @static
         * @param {blerpc.IDataWriteRequest=} [properties] Properties to set
         * @returns {blerpc.DataWriteRequest} DataWriteRequest instance
         */
        DataWriteRequest.create = function create(properties) {
            return new DataWriteRequest(properties);
        };

        /**
         * Encodes the specified DataWriteRequest message. Does not implicitly {@link blerpc.DataWriteRequest.verify|verify} messages.
         * @function encode
         * @memberof blerpc.DataWriteRequest
         * @static
         * @param {blerpc.IDataWriteRequest} message DataWriteRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        DataWriteRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.data != null && Object.hasOwnProperty.call(message, "data"))
                writer.uint32(/* id 1, wireType 2 =*/10).bytes(message.data);
            return writer;
        };

        /**
         * Encodes the specified DataWriteRequest message, length delimited. Does not implicitly {@link blerpc.DataWriteRequest.verify|verify} messages.
         * @function encodeDelimited
         * @memberof blerpc.DataWriteRequest
         * @static
         * @param {blerpc.IDataWriteRequest} message DataWriteRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        DataWriteRequest.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a DataWriteRequest message from the specified reader or buffer.
         * @function decode
         * @memberof blerpc.DataWriteRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {blerpc.DataWriteRequest} DataWriteRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        DataWriteRequest.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            var end = length === undefined ? reader.len : reader.pos + length, message = new $root.blerpc.DataWriteRequest();
            while (reader.pos < end) {
                var tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.data = reader.bytes();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a DataWriteRequest message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof blerpc.DataWriteRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {blerpc.DataWriteRequest} DataWriteRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        DataWriteRequest.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a DataWriteRequest message.
         * @function verify
         * @memberof blerpc.DataWriteRequest
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        DataWriteRequest.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.data != null && message.hasOwnProperty("data"))
                if (!(message.data && typeof message.data.length === "number" || $util.isString(message.data)))
                    return "data: buffer expected";
            return null;
        };

        /**
         * Creates a DataWriteRequest message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof blerpc.DataWriteRequest
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {blerpc.DataWriteRequest} DataWriteRequest
         */
        DataWriteRequest.fromObject = function fromObject(object) {
            if (object instanceof $root.blerpc.DataWriteRequest)
                return object;
            var message = new $root.blerpc.DataWriteRequest();
            if (object.data != null)
                if (typeof object.data === "string")
                    $util.base64.decode(object.data, message.data = $util.newBuffer($util.base64.length(object.data)), 0);
                else if (object.data.length >= 0)
                    message.data = object.data;
            return message;
        };

        /**
         * Creates a plain object from a DataWriteRequest message. Also converts values to other types if specified.
         * @function toObject
         * @memberof blerpc.DataWriteRequest
         * @static
         * @param {blerpc.DataWriteRequest} message DataWriteRequest
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        DataWriteRequest.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            var object = {};
            if (options.defaults)
                if (options.bytes === String)
                    object.data = "";
                else {
                    object.data = [];
                    if (options.bytes !== Array)
                        object.data = $util.newBuffer(object.data);
                }
            if (message.data != null && message.hasOwnProperty("data"))
                object.data = options.bytes === String ? $util.base64.encode(message.data, 0, message.data.length) : options.bytes === Array ? Array.prototype.slice.call(message.data) : message.data;
            return object;
        };

        /**
         * Converts this DataWriteRequest to JSON.
         * @function toJSON
         * @memberof blerpc.DataWriteRequest
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        DataWriteRequest.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for DataWriteRequest
         * @function getTypeUrl
         * @memberof blerpc.DataWriteRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        DataWriteRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/blerpc.DataWriteRequest";
        };

        return DataWriteRequest;
    })();

    blerpc.DataWriteResponse = (function() {

        /**
         * Properties of a DataWriteResponse.
         * @memberof blerpc
         * @interface IDataWriteResponse
         * @property {number|null} [length] DataWriteResponse length
         */

        /**
         * Constructs a new DataWriteResponse.
         * @memberof blerpc
         * @classdesc Represents a DataWriteResponse.
         * @implements IDataWriteResponse
         * @constructor
         * @param {blerpc.IDataWriteResponse=} [properties] Properties to set
         */
        function DataWriteResponse(properties) {
            if (properties)
                for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * DataWriteResponse length.
         * @member {number} length
         * @memberof blerpc.DataWriteResponse
         * @instance
         */
        DataWriteResponse.prototype.length = 0;

        /**
         * Creates a new DataWriteResponse instance using the specified properties.
         * @function create
         * @memberof blerpc.DataWriteResponse
         * @static
         * @param {blerpc.IDataWriteResponse=} [properties] Properties to set
         * @returns {blerpc.DataWriteResponse} DataWriteResponse instance
         */
        DataWriteResponse.create = function create(properties) {
            return new DataWriteResponse(properties);
        };

        /**
         * Encodes the specified DataWriteResponse message. Does not implicitly {@link blerpc.DataWriteResponse.verify|verify} messages.
         * @function encode
         * @memberof blerpc.DataWriteResponse
         * @static
         * @param {blerpc.IDataWriteResponse} message DataWriteResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        DataWriteResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.length != null && Object.hasOwnProperty.call(message, "length"))
                writer.uint32(/* id 1, wireType 0 =*/8).uint32(message.length);
            return writer;
        };

        /**
         * Encodes the specified DataWriteResponse message, length delimited. Does not implicitly {@link blerpc.DataWriteResponse.verify|verify} messages.
         * @function encodeDelimited
         * @memberof blerpc.DataWriteResponse
         * @static
         * @param {blerpc.IDataWriteResponse} message DataWriteResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        DataWriteResponse.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a DataWriteResponse message from the specified reader or buffer.
         * @function decode
         * @memberof blerpc.DataWriteResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {blerpc.DataWriteResponse} DataWriteResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        DataWriteResponse.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            var end = length === undefined ? reader.len : reader.pos + length, message = new $root.blerpc.DataWriteResponse();
            while (reader.pos < end) {
                var tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.length = reader.uint32();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a DataWriteResponse message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof blerpc.DataWriteResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {blerpc.DataWriteResponse} DataWriteResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        DataWriteResponse.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a DataWriteResponse message.
         * @function verify
         * @memberof blerpc.DataWriteResponse
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        DataWriteResponse.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.length != null && message.hasOwnProperty("length"))
                if (!$util.isInteger(message.length))
                    return "length: integer expected";
            return null;
        };

        /**
         * Creates a DataWriteResponse message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof blerpc.DataWriteResponse
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {blerpc.DataWriteResponse} DataWriteResponse
         */
        DataWriteResponse.fromObject = function fromObject(object) {
            if (object instanceof $root.blerpc.DataWriteResponse)
                return object;
            var message = new $root.blerpc.DataWriteResponse();
            if (object.length != null)
                message.length = object.length >>> 0;
            return message;
        };

        /**
         * Creates a plain object from a DataWriteResponse message. Also converts values to other types if specified.
         * @function toObject
         * @memberof blerpc.DataWriteResponse
         * @static
         * @param {blerpc.DataWriteResponse} message DataWriteResponse
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        DataWriteResponse.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            var object = {};
            if (options.defaults)
                object.length = 0;
            if (message.length != null && message.hasOwnProperty("length"))
                object.length = message.length;
            return object;
        };

        /**
         * Converts this DataWriteResponse to JSON.
         * @function toJSON
         * @memberof blerpc.DataWriteResponse
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        DataWriteResponse.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for DataWriteResponse
         * @function getTypeUrl
         * @memberof blerpc.DataWriteResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        DataWriteResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/blerpc.DataWriteResponse";
        };

        return DataWriteResponse;
    })();

    blerpc.CounterStreamRequest = (function() {

        /**
         * Properties of a CounterStreamRequest.
         * @memberof blerpc
         * @interface ICounterStreamRequest
         * @property {number|null} [count] CounterStreamRequest count
         */

        /**
         * Constructs a new CounterStreamRequest.
         * @memberof blerpc
         * @classdesc Represents a CounterStreamRequest.
         * @implements ICounterStreamRequest
         * @constructor
         * @param {blerpc.ICounterStreamRequest=} [properties] Properties to set
         */
        function CounterStreamRequest(properties) {
            if (properties)
                for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * CounterStreamRequest count.
         * @member {number} count
         * @memberof blerpc.CounterStreamRequest
         * @instance
         */
        CounterStreamRequest.prototype.count = 0;

        /**
         * Creates a new CounterStreamRequest instance using the specified properties.
         * @function create
         * @memberof blerpc.CounterStreamRequest
         * @static
         * @param {blerpc.ICounterStreamRequest=} [properties] Properties to set
         * @returns {blerpc.CounterStreamRequest} CounterStreamRequest instance
         */
        CounterStreamRequest.create = function create(properties) {
            return new CounterStreamRequest(properties);
        };

        /**
         * Encodes the specified CounterStreamRequest message. Does not implicitly {@link blerpc.CounterStreamRequest.verify|verify} messages.
         * @function encode
         * @memberof blerpc.CounterStreamRequest
         * @static
         * @param {blerpc.ICounterStreamRequest} message CounterStreamRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        CounterStreamRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.count != null && Object.hasOwnProperty.call(message, "count"))
                writer.uint32(/* id 1, wireType 0 =*/8).uint32(message.count);
            return writer;
        };

        /**
         * Encodes the specified CounterStreamRequest message, length delimited. Does not implicitly {@link blerpc.CounterStreamRequest.verify|verify} messages.
         * @function encodeDelimited
         * @memberof blerpc.CounterStreamRequest
         * @static
         * @param {blerpc.ICounterStreamRequest} message CounterStreamRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        CounterStreamRequest.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a CounterStreamRequest message from the specified reader or buffer.
         * @function decode
         * @memberof blerpc.CounterStreamRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {blerpc.CounterStreamRequest} CounterStreamRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        CounterStreamRequest.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            var end = length === undefined ? reader.len : reader.pos + length, message = new $root.blerpc.CounterStreamRequest();
            while (reader.pos < end) {
                var tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.count = reader.uint32();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a CounterStreamRequest message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof blerpc.CounterStreamRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {blerpc.CounterStreamRequest} CounterStreamRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        CounterStreamRequest.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a CounterStreamRequest message.
         * @function verify
         * @memberof blerpc.CounterStreamRequest
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        CounterStreamRequest.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.count != null && message.hasOwnProperty("count"))
                if (!$util.isInteger(message.count))
                    return "count: integer expected";
            return null;
        };

        /**
         * Creates a CounterStreamRequest message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof blerpc.CounterStreamRequest
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {blerpc.CounterStreamRequest} CounterStreamRequest
         */
        CounterStreamRequest.fromObject = function fromObject(object) {
            if (object instanceof $root.blerpc.CounterStreamRequest)
                return object;
            var message = new $root.blerpc.CounterStreamRequest();
            if (object.count != null)
                message.count = object.count >>> 0;
            return message;
        };

        /**
         * Creates a plain object from a CounterStreamRequest message. Also converts values to other types if specified.
         * @function toObject
         * @memberof blerpc.CounterStreamRequest
         * @static
         * @param {blerpc.CounterStreamRequest} message CounterStreamRequest
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        CounterStreamRequest.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            var object = {};
            if (options.defaults)
                object.count = 0;
            if (message.count != null && message.hasOwnProperty("count"))
                object.count = message.count;
            return object;
        };

        /**
         * Converts this CounterStreamRequest to JSON.
         * @function toJSON
         * @memberof blerpc.CounterStreamRequest
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        CounterStreamRequest.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for CounterStreamRequest
         * @function getTypeUrl
         * @memberof blerpc.CounterStreamRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        CounterStreamRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/blerpc.CounterStreamRequest";
        };

        return CounterStreamRequest;
    })();

    blerpc.CounterStreamResponse = (function() {

        /**
         * Properties of a CounterStreamResponse.
         * @memberof blerpc
         * @interface ICounterStreamResponse
         * @property {number|null} [seq] CounterStreamResponse seq
         * @property {number|null} [value] CounterStreamResponse value
         */

        /**
         * Constructs a new CounterStreamResponse.
         * @memberof blerpc
         * @classdesc Represents a CounterStreamResponse.
         * @implements ICounterStreamResponse
         * @constructor
         * @param {blerpc.ICounterStreamResponse=} [properties] Properties to set
         */
        function CounterStreamResponse(properties) {
            if (properties)
                for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * CounterStreamResponse seq.
         * @member {number} seq
         * @memberof blerpc.CounterStreamResponse
         * @instance
         */
        CounterStreamResponse.prototype.seq = 0;

        /**
         * CounterStreamResponse value.
         * @member {number} value
         * @memberof blerpc.CounterStreamResponse
         * @instance
         */
        CounterStreamResponse.prototype.value = 0;

        /**
         * Creates a new CounterStreamResponse instance using the specified properties.
         * @function create
         * @memberof blerpc.CounterStreamResponse
         * @static
         * @param {blerpc.ICounterStreamResponse=} [properties] Properties to set
         * @returns {blerpc.CounterStreamResponse} CounterStreamResponse instance
         */
        CounterStreamResponse.create = function create(properties) {
            return new CounterStreamResponse(properties);
        };

        /**
         * Encodes the specified CounterStreamResponse message. Does not implicitly {@link blerpc.CounterStreamResponse.verify|verify} messages.
         * @function encode
         * @memberof blerpc.CounterStreamResponse
         * @static
         * @param {blerpc.ICounterStreamResponse} message CounterStreamResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        CounterStreamResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.seq != null && Object.hasOwnProperty.call(message, "seq"))
                writer.uint32(/* id 1, wireType 0 =*/8).uint32(message.seq);
            if (message.value != null && Object.hasOwnProperty.call(message, "value"))
                writer.uint32(/* id 2, wireType 0 =*/16).int32(message.value);
            return writer;
        };

        /**
         * Encodes the specified CounterStreamResponse message, length delimited. Does not implicitly {@link blerpc.CounterStreamResponse.verify|verify} messages.
         * @function encodeDelimited
         * @memberof blerpc.CounterStreamResponse
         * @static
         * @param {blerpc.ICounterStreamResponse} message CounterStreamResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        CounterStreamResponse.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a CounterStreamResponse message from the specified reader or buffer.
         * @function decode
         * @memberof blerpc.CounterStreamResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {blerpc.CounterStreamResponse} CounterStreamResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        CounterStreamResponse.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            var end = length === undefined ? reader.len : reader.pos + length, message = new $root.blerpc.CounterStreamResponse();
            while (reader.pos < end) {
                var tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.seq = reader.uint32();
                        break;
                    }
                case 2: {
                        message.value = reader.int32();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a CounterStreamResponse message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof blerpc.CounterStreamResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {blerpc.CounterStreamResponse} CounterStreamResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        CounterStreamResponse.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a CounterStreamResponse message.
         * @function verify
         * @memberof blerpc.CounterStreamResponse
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        CounterStreamResponse.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.seq != null && message.hasOwnProperty("seq"))
                if (!$util.isInteger(message.seq))
                    return "seq: integer expected";
            if (message.value != null && message.hasOwnProperty("value"))
                if (!$util.isInteger(message.value))
                    return "value: integer expected";
            return null;
        };

        /**
         * Creates a CounterStreamResponse message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof blerpc.CounterStreamResponse
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {blerpc.CounterStreamResponse} CounterStreamResponse
         */
        CounterStreamResponse.fromObject = function fromObject(object) {
            if (object instanceof $root.blerpc.CounterStreamResponse)
                return object;
            var message = new $root.blerpc.CounterStreamResponse();
            if (object.seq != null)
                message.seq = object.seq >>> 0;
            if (object.value != null)
                message.value = object.value | 0;
            return message;
        };

        /**
         * Creates a plain object from a CounterStreamResponse message. Also converts values to other types if specified.
         * @function toObject
         * @memberof blerpc.CounterStreamResponse
         * @static
         * @param {blerpc.CounterStreamResponse} message CounterStreamResponse
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        CounterStreamResponse.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            var object = {};
            if (options.defaults) {
                object.seq = 0;
                object.value = 0;
            }
            if (message.seq != null && message.hasOwnProperty("seq"))
                object.seq = message.seq;
            if (message.value != null && message.hasOwnProperty("value"))
                object.value = message.value;
            return object;
        };

        /**
         * Converts this CounterStreamResponse to JSON.
         * @function toJSON
         * @memberof blerpc.CounterStreamResponse
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        CounterStreamResponse.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for CounterStreamResponse
         * @function getTypeUrl
         * @memberof blerpc.CounterStreamResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        CounterStreamResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/blerpc.CounterStreamResponse";
        };

        return CounterStreamResponse;
    })();

    blerpc.CounterUploadRequest = (function() {

        /**
         * Properties of a CounterUploadRequest.
         * @memberof blerpc
         * @interface ICounterUploadRequest
         * @property {number|null} [seq] CounterUploadRequest seq
         * @property {number|null} [value] CounterUploadRequest value
         */

        /**
         * Constructs a new CounterUploadRequest.
         * @memberof blerpc
         * @classdesc Represents a CounterUploadRequest.
         * @implements ICounterUploadRequest
         * @constructor
         * @param {blerpc.ICounterUploadRequest=} [properties] Properties to set
         */
        function CounterUploadRequest(properties) {
            if (properties)
                for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * CounterUploadRequest seq.
         * @member {number} seq
         * @memberof blerpc.CounterUploadRequest
         * @instance
         */
        CounterUploadRequest.prototype.seq = 0;

        /**
         * CounterUploadRequest value.
         * @member {number} value
         * @memberof blerpc.CounterUploadRequest
         * @instance
         */
        CounterUploadRequest.prototype.value = 0;

        /**
         * Creates a new CounterUploadRequest instance using the specified properties.
         * @function create
         * @memberof blerpc.CounterUploadRequest
         * @static
         * @param {blerpc.ICounterUploadRequest=} [properties] Properties to set
         * @returns {blerpc.CounterUploadRequest} CounterUploadRequest instance
         */
        CounterUploadRequest.create = function create(properties) {
            return new CounterUploadRequest(properties);
        };

        /**
         * Encodes the specified CounterUploadRequest message. Does not implicitly {@link blerpc.CounterUploadRequest.verify|verify} messages.
         * @function encode
         * @memberof blerpc.CounterUploadRequest
         * @static
         * @param {blerpc.ICounterUploadRequest} message CounterUploadRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        CounterUploadRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.seq != null && Object.hasOwnProperty.call(message, "seq"))
                writer.uint32(/* id 1, wireType 0 =*/8).uint32(message.seq);
            if (message.value != null && Object.hasOwnProperty.call(message, "value"))
                writer.uint32(/* id 2, wireType 0 =*/16).int32(message.value);
            return writer;
        };

        /**
         * Encodes the specified CounterUploadRequest message, length delimited. Does not implicitly {@link blerpc.CounterUploadRequest.verify|verify} messages.
         * @function encodeDelimited
         * @memberof blerpc.CounterUploadRequest
         * @static
         * @param {blerpc.ICounterUploadRequest} message CounterUploadRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        CounterUploadRequest.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a CounterUploadRequest message from the specified reader or buffer.
         * @function decode
         * @memberof blerpc.CounterUploadRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {blerpc.CounterUploadRequest} CounterUploadRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        CounterUploadRequest.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            var end = length === undefined ? reader.len : reader.pos + length, message = new $root.blerpc.CounterUploadRequest();
            while (reader.pos < end) {
                var tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.seq = reader.uint32();
                        break;
                    }
                case 2: {
                        message.value = reader.int32();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a CounterUploadRequest message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof blerpc.CounterUploadRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {blerpc.CounterUploadRequest} CounterUploadRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        CounterUploadRequest.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a CounterUploadRequest message.
         * @function verify
         * @memberof blerpc.CounterUploadRequest
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        CounterUploadRequest.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.seq != null && message.hasOwnProperty("seq"))
                if (!$util.isInteger(message.seq))
                    return "seq: integer expected";
            if (message.value != null && message.hasOwnProperty("value"))
                if (!$util.isInteger(message.value))
                    return "value: integer expected";
            return null;
        };

        /**
         * Creates a CounterUploadRequest message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof blerpc.CounterUploadRequest
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {blerpc.CounterUploadRequest} CounterUploadRequest
         */
        CounterUploadRequest.fromObject = function fromObject(object) {
            if (object instanceof $root.blerpc.CounterUploadRequest)
                return object;
            var message = new $root.blerpc.CounterUploadRequest();
            if (object.seq != null)
                message.seq = object.seq >>> 0;
            if (object.value != null)
                message.value = object.value | 0;
            return message;
        };

        /**
         * Creates a plain object from a CounterUploadRequest message. Also converts values to other types if specified.
         * @function toObject
         * @memberof blerpc.CounterUploadRequest
         * @static
         * @param {blerpc.CounterUploadRequest} message CounterUploadRequest
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        CounterUploadRequest.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            var object = {};
            if (options.defaults) {
                object.seq = 0;
                object.value = 0;
            }
            if (message.seq != null && message.hasOwnProperty("seq"))
                object.seq = message.seq;
            if (message.value != null && message.hasOwnProperty("value"))
                object.value = message.value;
            return object;
        };

        /**
         * Converts this CounterUploadRequest to JSON.
         * @function toJSON
         * @memberof blerpc.CounterUploadRequest
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        CounterUploadRequest.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for CounterUploadRequest
         * @function getTypeUrl
         * @memberof blerpc.CounterUploadRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        CounterUploadRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/blerpc.CounterUploadRequest";
        };

        return CounterUploadRequest;
    })();

    blerpc.CounterUploadResponse = (function() {

        /**
         * Properties of a CounterUploadResponse.
         * @memberof blerpc
         * @interface ICounterUploadResponse
         * @property {number|null} [receivedCount] CounterUploadResponse receivedCount
         */

        /**
         * Constructs a new CounterUploadResponse.
         * @memberof blerpc
         * @classdesc Represents a CounterUploadResponse.
         * @implements ICounterUploadResponse
         * @constructor
         * @param {blerpc.ICounterUploadResponse=} [properties] Properties to set
         */
        function CounterUploadResponse(properties) {
            if (properties)
                for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * CounterUploadResponse receivedCount.
         * @member {number} receivedCount
         * @memberof blerpc.CounterUploadResponse
         * @instance
         */
        CounterUploadResponse.prototype.receivedCount = 0;

        /**
         * Creates a new CounterUploadResponse instance using the specified properties.
         * @function create
         * @memberof blerpc.CounterUploadResponse
         * @static
         * @param {blerpc.ICounterUploadResponse=} [properties] Properties to set
         * @returns {blerpc.CounterUploadResponse} CounterUploadResponse instance
         */
        CounterUploadResponse.create = function create(properties) {
            return new CounterUploadResponse(properties);
        };

        /**
         * Encodes the specified CounterUploadResponse message. Does not implicitly {@link blerpc.CounterUploadResponse.verify|verify} messages.
         * @function encode
         * @memberof blerpc.CounterUploadResponse
         * @static
         * @param {blerpc.ICounterUploadResponse} message CounterUploadResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        CounterUploadResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.receivedCount != null && Object.hasOwnProperty.call(message, "receivedCount"))
                writer.uint32(/* id 1, wireType 0 =*/8).uint32(message.receivedCount);
            return writer;
        };

        /**
         * Encodes the specified CounterUploadResponse message, length delimited. Does not implicitly {@link blerpc.CounterUploadResponse.verify|verify} messages.
         * @function encodeDelimited
         * @memberof blerpc.CounterUploadResponse
         * @static
         * @param {blerpc.ICounterUploadResponse} message CounterUploadResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        CounterUploadResponse.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a CounterUploadResponse message from the specified reader or buffer.
         * @function decode
         * @memberof blerpc.CounterUploadResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {blerpc.CounterUploadResponse} CounterUploadResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        CounterUploadResponse.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            var end = length === undefined ? reader.len : reader.pos + length, message = new $root.blerpc.CounterUploadResponse();
            while (reader.pos < end) {
                var tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.receivedCount = reader.uint32();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a CounterUploadResponse message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof blerpc.CounterUploadResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {blerpc.CounterUploadResponse} CounterUploadResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        CounterUploadResponse.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a CounterUploadResponse message.
         * @function verify
         * @memberof blerpc.CounterUploadResponse
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        CounterUploadResponse.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.receivedCount != null && message.hasOwnProperty("receivedCount"))
                if (!$util.isInteger(message.receivedCount))
                    return "receivedCount: integer expected";
            return null;
        };

        /**
         * Creates a CounterUploadResponse message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof blerpc.CounterUploadResponse
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {blerpc.CounterUploadResponse} CounterUploadResponse
         */
        CounterUploadResponse.fromObject = function fromObject(object) {
            if (object instanceof $root.blerpc.CounterUploadResponse)
                return object;
            var message = new $root.blerpc.CounterUploadResponse();
            if (object.receivedCount != null)
                message.receivedCount = object.receivedCount >>> 0;
            return message;
        };

        /**
         * Creates a plain object from a CounterUploadResponse message. Also converts values to other types if specified.
         * @function toObject
         * @memberof blerpc.CounterUploadResponse
         * @static
         * @param {blerpc.CounterUploadResponse} message CounterUploadResponse
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        CounterUploadResponse.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            var object = {};
            if (options.defaults)
                object.receivedCount = 0;
            if (message.receivedCount != null && message.hasOwnProperty("receivedCount"))
                object.receivedCount = message.receivedCount;
            return object;
        };

        /**
         * Converts this CounterUploadResponse to JSON.
         * @function toJSON
         * @memberof blerpc.CounterUploadResponse
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        CounterUploadResponse.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for CounterUploadResponse
         * @function getTypeUrl
         * @memberof blerpc.CounterUploadResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        CounterUploadResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/blerpc.CounterUploadResponse";
        };

        return CounterUploadResponse;
    })();

    return blerpc;
})();

module.exports = $root;
